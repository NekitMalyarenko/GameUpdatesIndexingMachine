package db

import (
	"github.com/juju/errors"
	"upper.io/db.v3"
	"log"
)

type Update struct {
	Id          int64  `db:"id,omitempty" json:"id"`
	URl         string `db:"url"          json:"-"`
	OriginalURL string `db:"original_url" json:"original_url"`

	Date      string `db:"date"      json:"date"`
	Title     string `db:"title"     json:"title"`
	TitleImg *string `db:"title_img" json:"title_img"`
	ShortDes  string `db:"short_des" json:"short_des"`

	Views    int `db:"views"    json:"views"`
	Likes    int `db:"likes"    json:"likes"`
	Dislikes int `db:"dislikes" json:"dislikes"`
}


func GetUpdateById(id int64) (update *Update, err error) {
	err = errors.Trace(GetInstance().db.Collection("updates").Find(db.Cond{"id" : id}).One(&update))
	return
}


func (update *Update) InsertToDB() error {
	log.Println(update)
	_, err := GetInstance().db.InsertInto("updates").Values(update).Exec()
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(GetInstance().db.Collection("updates").Find().
		OrderBy("-id").Limit(1).One(update))
}


func (update *Update) SaveToDB() error {
	return errors.Trace(updateInDB("updates", db.Cond{"id": update.Id}, update))
}