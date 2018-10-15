package db

import (
	"upper.io/db.v3/postgresql"
	"github.com/juju/errors"
	"upper.io/db.v3"
)

type Game struct {
	Id            int64            `db:"id,omitempty"`
	Name          string           `db:"name"`
	ShortName     string           `db:"short_name"`
	LastUpdatesId postgresql.JSONB `db:"last_updates_id"`
	SupportedLang postgresql.JSONB `db:"supported_lang"`
}


func GetAllGames() (games []*Game, err error) {
	err = errors.Trace(getFromDB("games", db.Cond{}).All(&games))
	return
}


func (game *Game) SaveToDB() error {
	return errors.Trace(updateInDB("games", db.Cond{"id": game.Id}, game))
}


func (game *Game) GetSupportedLang() []string {
	res := make([]string, 0)

	for _, val := range game.SupportedLang.V.([]interface{}) {
		res = append(res, val.(string))
	}

	return res
}


func (game *Game) UpdateLastUpdateId(lang, newUpdateId string) {
	raw := game.LastUpdatesId.V.(map[string]interface{})
	raw[lang] = newUpdateId
}


func (game *Game) GetLastUpdateId(lang string) string {
	raw := game.LastUpdatesId.V.(map[string]interface{})

	for key, val := range raw {
		if key == lang {
			return val.(string)
		}
	}

	return ""
}


func (game *Game) GetLastUpdatesIdLen() int {
	return len(game.LastUpdatesId.V.(map[string]interface{}))
}
