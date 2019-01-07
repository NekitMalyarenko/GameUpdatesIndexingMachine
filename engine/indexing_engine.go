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
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/services/image_service"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/images"
)

import _ "image/jpeg"
import (
	_ "image/png"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/services/firebase_notifications"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/firebase_topics"
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

					updateHTML, err = parseImages(update, updateHTML, lastUpdateId, language, game.Id)
					if err != nil {
						log.Println(errors.Details(err))
						goto exit
					}
					log.Println("\t\tsuccessfully parsed images")

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
						wrapper = new(db.UpdateWrapper)
						wrapper.GameId = game.Id
						wrapper.UpdateId = lastUpdateId
						wrapper.Data.V = make(map[string]interface{})
						wrapper.AddLastUpdates(language, update.Id)
						err = wrapper.InsertToDB()
						if err != nil {
							log.Println(errors.Details(err))
							goto exit
						}
						log.Println("\t\tsuccesfully created new wrapper")
					}

					if language == languages.EN {
						topic, ok := firebase_topics.Topics[int(game.Id)]
						if ok  {
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

	case games.Fortnite:
		return getLastFortniteUpdateId(lang)

	default:
		return
	}
}


func getUpdate(gameId int64, url string) (update *db.Update, updateHTML string, err error) {

	switch gameId {

	case games.CsgoBlog:
		return getCSGOBlogUpdate(url)

	case games.Fortnite:
		return getFortniteUpdate(url)

	default:
		return
	}
}


func compare(gameId int64, updateId1, updateId2 string) bool {

	switch gameId {

	case games.CsgoBlog:
		return compareCSGOBlog(updateId1, updateId2)

	case games.Fortnite:
		return updateId1 != updateId2

	default:
		return false
	}
}


func parseImages(update *db.Update, updateHTML, lastUpdateId, language string, gameId int64) (string, error) {
	updateDoc, Err := goquery.NewDocumentFromReader(strings.NewReader(updateHTML))
	if Err != nil {
		return "", errors.Trace(Err)
	}

	var (
		url string
		ok  bool
	)


	updateDoc.Find("img").Each(func(imageIndex int, s *goquery.Selection) {
		if Err != nil {
			return
		}

		if language == languages.EN {
			url, ok = s.Attr("src")
			if ok {
				img, err := image_service.DownloadImage(url)
				if err != nil {
					Err = errors.Trace(err)
					return
				}

				resizedImg, err := image_service.ResizeImage(img, images.Width, images.Height)
				if err != nil {
					Err = errors.Trace(err)
					return
				}

				_, err = cloudinary.SaveImage(resizedImg, gameId, lastUpdateId, int64(imageIndex + 1))
				if err != nil {
					Err = errors.Trace(err)
					return
				}
			}
		}

		// update id need to be string like "12.12.2018"
		newImageUrl := "/getImage/"+ strconv.FormatInt(gameId, 10) +
			"/" + lastUpdateId + "/" + strconv.Itoa(imageIndex + 1)
		s.SetAttr("src", newImageUrl)

		if update.TitleImg == url {
			update.TitleImg = newImageUrl
		}
	})
	if Err != nil {
		return "", Err
	}

	if !strings.Contains(update.TitleImg, "/getImage/" +
		strconv.FormatInt(gameId, 10) + "/"+ lastUpdateId + "/") {
		img, err := image_service.DownloadImage(update.TitleImg)
		if err != nil {
			return "", errors.Trace(err)
		}

		resizedImg, err := image_service.ResizeImage(img, images.Width, images.Height)
		if err != nil {
			return "", errors.Trace(err)
		}

		_, err = cloudinary.SaveImage(resizedImg, gameId, lastUpdateId,0)
		if err != nil {
			return "", errors.Trace(err)
		}

		update.TitleImg =  "/getImage/"+ strconv.FormatInt(gameId, 10) +
			"/" + lastUpdateId + "/0"
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