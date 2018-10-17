package engine

import (
	"github.com/juju/errors"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"unicode/utf8"

	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/languages"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/db"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/articles"
)


func getLastCSGOBlogUpdateId(lang string) (id string, url string, err error) {
	var (
		ok = false
		downloadURL = ""
	)

	if lang == languages.EN {
		downloadURL = "http://blog.counter-strike.net/"
	} else {
		downloadURL = "http://blog.counter-strike.net/" + lang
	}

	page, err := downloadPage(downloadURL)
	if err != nil {
		return "", "", err
	}

	id = page.Find(".inner_post").Eq(0).Find(".post_date").Text()
	id = strings.Replace(id, " ", "", -1)
	id = string([]byte(id)[:len(id) - 1])

	url, ok = page.Find(".inner_post").Eq(0).Find("h2 a").First().Attr("href")
	if ok == false {
		errors.Trace(errors.New("can't get url"))
		return "", "", err
	}

	return
}


func getCSGOBlogUpdate(url string) (update *db.Update, updateHTML string, _ error) {
	var res = ""

	page, err := downloadPage(url)
	if err != nil {
		return nil, "", errors.Trace(err)
	}

	update = new(db.Update)

	update.Title = page.Find(".inner_post").Eq(0).Find("h2 a").First().Text()

	update.Date = page.Find(".inner_post").Eq(0).Find(".post_date").Text()
	update.Date = strings.Replace(update.Date, " ", "", -1)
	update.Date = string([]byte(update.Date)[:len(update.Date) - 1])

	update.OriginalURL = url

	page.Find(".inner_post p script").Remove()
	page.Find(".inner_post p").Each(func(i int, selection *goquery.Selection) {
		if i == 0 || err != nil{
			return
		}

		res, err = selection.Html()
		if err != nil {
			err = errors.Trace(err)
		}
		updateHTML += "<p>" + res + "</p>"

		if utf8.RuneCountInString(update.ShortDes) == articles.ShortDesLength - 1 {
			return
		}  else {
			update.ShortDes += selection.Text()

			if utf8.RuneCountInString(update.ShortDes) > articles.ShortDesLength - 1 {
				update.ShortDes = string([]byte(update.ShortDes)[:articles.ShortDesLength - 3]) + "..."
			}
		}
	})

	if page.Find(".inner_post p img").Size() > 1 {
		raw, _ := page.Find(".inner_post img").Eq(1).Attr("src")
		update.TitleImg = &raw
	} else {
		update.TitleImg = nil
	}

	return update, updateHTML, errors.Trace(err)
}


func compareCSGOBlog(updateId1, updateId2 string) bool {
	return updateId1 > updateId2
}
