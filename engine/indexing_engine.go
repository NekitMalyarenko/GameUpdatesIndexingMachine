package engine

import (
	"log"
	"time"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"github.com/juju/errors"
	"strconv"
	"strings"

	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/db"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/services/cloudinary"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/games"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/languages"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/services/firebase_notifications"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/firebase_topics"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/services/image_service"
	"image"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/images"
)

var (
	parseRUString = func(input string) string {
		return strings.Replace(input, "о", "е", -1)
	}
)


func Start() {
	dbGames, err := db.GetAllGames()
	if err != nil {
		log.Println(err)
		return
	}

	var (
		supportedLang []string

		minutes int
		seconds int

		minutes1 int
		seconds1 int

		empty bool
	)

	for {

		for _, game := range dbGames {
			log.Println("starting indexing", game.ShortName, "...")
			minutes = time.Now().Minute()
			seconds = time.Now().Second()

			supportedLang = game.GetSupportedLang()
			empty = game.GetLastUpdatesIdLen() == 0

			for _, language := range supportedLang {
				log.Println("\t", language, "...")
				minutes1 = time.Now().Minute()
				seconds1 = time.Now().Second()
				hasError := false

				lastUpdateId, url, err := getLastUpdateId(game.Id, language)
				if err != nil {
					log.Println(errors.Details(err))
					hasError = true
				}

				if hasError {
					goto exit
				} else {
					log.Println("\t\tlatesUpdateId(web):", lastUpdateId, "lastUpdateId(DB):", game.GetLastUpdateId(language))
				}

				if empty || len(game.GetLastUpdateId(language)) == 0 {
					log.Println("\t\t first update")
					goto saveUpdate
				} else if compare(game.Id, lastUpdateId, game.GetLastUpdateId(language)) {
					log.Println("\t\t new update")
					goto saveUpdate
				} else {
					log.Println("\t\t no new updates", )
					goto exit
				}

			saveUpdate:
				{
					update, updateHTML, err := getUpdate(game.Id, url)
					if err != nil {
						log.Println(errors.Details(err))
						goto exit
					}
					log.Println("\t\tsuccesfully got update")

					/*updateHTML, err = parseImages(updateHTML, lastUpdateId, language, game.Id)
					if err != nil {
						log.Println(errors.Details(err))
						goto exit
					}
					log.Println("\t\t successfully parsed images")*/

					//parseAllStringResources(update, language)

					update.URl, err = cloudinary.Save(game.Id, lastUpdateId, language, updateHTML)
					if err != nil {
						log.Println(errors.Details(err))
						goto exit
					}
					log.Println("\t\tsuccesfully uploaded on cloudinary")

					err = update.InsertToDB()
					if err != nil {
						log.Println(errors.Details(err))
						goto exit
					}
					log.Println("\t\tsuccesfully inserted update in db with id:", update.Id)

					game.UpdateLastUpdateId(language, lastUpdateId)
					if game.SaveToDB() != nil {
						log.Println(errors.Details(err))
						goto exit
					}
					log.Println("\t\tsuccesfully updated game latestUpdateId")

					wrapper, err := db.GetUpdateWrapper(game.Id, lastUpdateId)
					if err != nil {
						log.Println(errors.Details(err))
						goto exit
					}

					if wrapper != nil {
						wrapper.AddLastUpdates(language, update.Id)
						wrapper.SaveToDB()
						log.Println("\t\tsuccesfully updated wrapper")
					} else {
						wrapper := new(db.UpdateWrapper)
						wrapper.GameId = game.Id
						wrapper.UpdateId = lastUpdateId
						wrapper.Data.V = make(map[string]interface{})
						wrapper.AddLastUpdates(language, update.Id)
						wrapper.InsertToDB()
						log.Println("\t\tsuccesfully created new wrapper")
					}

					topic, ok := firebase_topics.Topics[int(game.Id)]
					if ok {
						notification := firebase_notifications.NotificationData{
							Topic: topic,
							Body: update.Title,
							Title: "New update in " + game.ShortName,
							Id: strconv.FormatInt(wrapper.Id, 10),
						}
						err = notification.Send()
						if err != nil {
							log.Println(errors.Details(errors.Trace(err)))
						}
					} else {
						log.Println("\t\ttopic not found for", game.Id)
					}
				}

			exit:
				{
					log.Println("\t", language, "took:", time.Now().Minute()-minutes1,
						"minutes", time.Now().Second()-seconds1, "seconds")
				}

			}

			log.Println("indexing for", game.ShortName, "took:", time.Now().Minute() - minutes,
				"minutes", time.Now().Second() - seconds, "seconds")
		}

		time.Sleep(1 * time.Minute)
	}

}


func getLastUpdateId(gameId int64, lang string) (id string, url string, err error) {

	switch gameId {

	case games.CsgoBlog:
		return getLastCSGOBlogUpdateId(lang)

	default:
		return
	}
}


func getUpdate(gameId int64, url string) (update *db.Update, updateHTML string, err error) {

	switch gameId {

	case games.CsgoBlog:
		return getCSGOBlogUpdate(url)

	default:
		return
	}
}


func compare(gameId int64, updateId1, updateId2 string) bool {

	switch gameId {

	case games.CsgoBlog:
		return compareCSGOBlog(updateId1, updateId2)

	default:
		return false
	}
}


func parseImages(updateHTML, lastUpdateId, language string, gameId int64) (string, error){
	updateDoc, Err := goquery.NewDocumentFromReader(strings.NewReader(updateHTML))
	if Err != nil {
		return "", errors.Trace(Err)
	}

	updateDoc.Find("img").Each(func(imageIndex int, s *goquery.Selection) {
		if Err != nil {
			return
		}

		if language == languages.EN {
			url, _ := s.Attr("src")

			img, err := downloadImage(url)
			if err != nil {
				Err = errors.Trace(err)
				return
			}

			resizedImg, err := services.ResizeImage(img, images.Width, images.Height)
			if err != nil {
				Err = errors.Trace(err)
				return
			}

			_, err = cloudinary.SaveImage(resizedImg, gameId, lastUpdateId, int64(imageIndex))
			if err != nil {
				Err = errors.Trace(err)
				return
			}
		}

		// update id need to be string like "12.12.2018"
		newImageUrl := "/get-image/"+ strconv.FormatInt(gameId, 10) +
			"/" + lastUpdateId + "/" + strconv.Itoa(imageIndex)
		s.SetAttr("src", newImageUrl)
	})
	if Err != nil {
		return "", Err
	}

	updateHTML, Err = updateDoc.Html()
	if Err != nil {
		return "", errors.Trace(Err)
	}

	return updateHTML[25:len(updateHTML)-14], nil
}


func parseAllStringResources(update *db.Update, lang string) {

	switch lang {

	case languages.RU:
		update.Title = parseRUString(update.Title)
		log.Println("shortDes old:", update.ShortDes)
		update.ShortDes = parseRUString(update.ShortDes)
		log.Println("shortDes new:", update.ShortDes)

	}

}


func downloadPage(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Trace(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New("error status code:" + strconv.Itoa(res.StatusCode))
	}

	return goquery.NewDocumentFromReader(res.Body)
}


func downloadImage(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer response.Body.Close()

	img, _, err := image.Decode(response.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return img, nil
}