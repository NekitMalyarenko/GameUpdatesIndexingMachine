package engine

import (
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/db"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/languages"
	"github.com/juju/errors"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"unicode/utf8"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/const/articles"
	"golang.org/x/net/html"
	"time"
	"log"
)


func getLastFortniteUpdateId(lang string) (id string, url string, _ error) {
	var ok bool

	if lang == languages.EN {
		lang = "en-US"
	}
	downloadURL := "https://www.epicgames.com/fortnite/" + lang + "/news"

	page, err := downloadPage(downloadURL)
	if err != nil {
		return "", "", err
	}

	root := page.Find(".top-featured-activity").Eq(0).Find("a").Eq(0)

	url, ok = root.Attr("href")
	if !ok {
		return "", "", errors.Trace(errors.New("href is empty"))
	}
	url = "https://www.epicgames.com" + url

	pattern := "/fortnite/" + lang + "/"
	id = url[strings.Index(url, pattern) + len(pattern):]
	id = strings.Replace(id, "/", ".", -1)

	return
}


func getFortniteUpdate(url string) (update *db.Update, updateHTML string, _ error) {
	log.Println(url)
	page, err := downloadPage(url)
	if err != nil {
		return nil, "", err
	}

	update = new(db.Update)
	update.OriginalURL = url

	if strings.Contains(url ,"news") {
		update.Title = page.Find(".blog-header-info .blog-header-title").Text()

		layout := func() string {
			if strings.Contains(url, "/en-US/") {
				return "02.01.2016"
			} else {
				return "01.02.2016"
			}
		}()

		rawDate, err := time.Parse(layout,  page.Find(".blog-header-info .blog-header-date").Text())
		if err != nil {
			return nil, "", errors.Trace(err)
		}

		update.Date = rawDate.Format("2006-01-02")

		img, has := page.Find(".blog-header img").Attr("src")
		if has {
			update.TitleImg = img
		} else {
			update.TitleImg = ""
		}

		page.Find("#cmsSection style").Remove()

		//log.Println("cms:", page.Find(".cmsSection").Text())
		page.Find("#cmsSection").Children().Each(func(i int, selection *goquery.Selection) {
			if err != nil {
				return
			}

			if selection.Nodes[0].Data != "img" && selection.Nodes[0].Data != "a" {
				selection.Nodes[0].Attr = make([]html.Attribute, 0)
			}

			if utf8.RuneCountInString(update.ShortDes) == articles.ShortDesLength {
				return
			}  else {
				update.ShortDes += selection.Text()
				update.ShortDes = strings.Replace(update.ShortDes, "\n", "", -1)
				//log.Println("after adding length:", utf8.RuneCountInString(shortDes))

				//log.Println("i:", i, "text:", selection.Text())

				if utf8.RuneCountInString(update.ShortDes) > articles.ShortDesLength {
					update.ShortDes = string([]rune(update.ShortDes)[:articles.ShortDesLength - 3]) + "..."
				}
			}
		})
		if err != nil {
			return nil, "", errors.Trace(err)
		}

		if utf8.RuneCountInString(update.ShortDes) < articles.ShortDesLength {
			update.ShortDes = page.Find("#cmsSection").Text()
			update.ShortDes = strings.Replace(update.ShortDes, "\n", "", -1)

			if utf8.RuneCountInString(update.ShortDes) > articles.ShortDesLength {
				update.ShortDes = string([]rune(update.ShortDes)[:articles.ShortDesLength - 3]) + "..."
			}
		}

		updateHTML, err = page.Find("#cmsSection").Html()
		if err != nil {
			return nil, "", errors.Trace(err)
		}

	} else if strings.Contains(url, "patch-notes") {
		update.Title = page.Find(".patch-notes-navigation .patch-container").Children().Eq(0).Children().Eq(0).Text()
		update.Date = time.Now().Format("2006-01-02")

		style, has := page.Find(".background-image").Attr("style")
		if has && strings.Index(style, "url(") != -1 && strings.Index(style, ")") != -1 {
			res := style[strings.Index(style, "url(")+4 : strings.Index(style, ")") ]
			update.TitleImg = res
		} else {
			update.TitleImg = ""
		}

		page.Find(".patch-notes-text style").Remove()

		page.Find(".patch-notes-text").Each(func(i int, selection *goquery.Selection) {
			if err != nil {
				return
			}

			//log.Println("i:", i)

			updateHTML += "<h1>" + selection.Find(".row").Children().Eq(0).Find("h1").Text() + "</h1>"

			temp := selection.Find(".patch-notes-description")
			innerHTML, Err := temp.Html()
			if Err != nil {
				err = errors.Trace(Err)
			}

			updateHTML += "<" + temp.Nodes[0].Data + ">" + innerHTML + "</" + temp.Nodes[0].Data + ">"

			log.Println(temp.Nodes[0].Data)

			if utf8.RuneCountInString(update.ShortDes) == articles.ShortDesLength {
				return
			}  else if temp.Nodes[0].Data != "style"{
				update.ShortDes += temp.Text()
				if strings.Index(update.ShortDes, "<") != -1 && strings.Index(update.ShortDes, ">") != 1 {
					update.ShortDes = update.ShortDes[strings.Index(update.ShortDes, "<"):] +
						update.ShortDes[:strings.Index(update.ShortDes, ">")]
				}
				update.ShortDes = strings.Replace(update.ShortDes, "\n", "", -1)
				//log.Println("after adding length:", utf8.RuneCountInString(shortDes))

				//log.Println("i:", i, "text:", selection.Text())

				if utf8.RuneCountInString(update.ShortDes) > articles.ShortDesLength {
					update.ShortDes = string([]rune(update.ShortDes)[:articles.ShortDesLength - 3]) + "..."
				}
			}

			//log.Println("tag:", selection.Nodes[0].Data)

			/*res, err = selection.Html()
			log.Println("test:", selection.Text())
			//log.Println(res)
			if err != nil {
				err = errors.Trace(err)
			}*/
			//updateHTML += "<p>" + res + "</p>"

			//log.Println("legnth:", utf8.RuneCountInString(shortDes), "max:", articles.ShortDesLength)
		})
		if err != nil {
			return nil, "", err
		}
	}

	return update, updateHTML, nil
}