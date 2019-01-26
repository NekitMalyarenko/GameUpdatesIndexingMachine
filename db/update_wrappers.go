package db

import (
	"upper.io/db.v3/postgresql"
	"upper.io/db.v3"
	"github.com/juju/errors"
)

type UpdateWrapper struct {
	Id       int64            `db:"id,omitempty"`
	UpdateId string           `db:"update_id"`
	GameId   int64			  `db:"game_id"`
	Data     postgresql.JSONB `db:"data"`
	Views    int              `db:"views"`
	Likes    postgresql.JSONB `db:"likes"`
	Dislikes postgresql.JSONB `db:"dislikes"`
}


func GetUpdateWrapper(gameId int64, updateId string) (*UpdateWrapper, error) {
	var wrapper *UpdateWrapper

	res := getFromDB("wrappers", db.Cond{"update_id" : updateId,
		"game_id" : gameId})

	exists, err := res.Exists()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if exists {
		res.One(&wrapper)
		return wrapper, nil
	} else {
		return nil, nil
	}
}


func GetUpdateWrapperById(id int64) (*UpdateWrapper, error) {
	var wrapper *UpdateWrapper

	res := getFromDB("wrappers", db.Cond{"id" : id})

	exists, err := res.Exists()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if exists {
		res.One(&wrapper)
		return wrapper, nil
	} else {
		return nil, nil
	}
}


func GetLastGameUpdateIdByLanguage(language string) (string, error) {
	row, err := GetInstance().db.QueryRow("select update_id from wrappers WHERE (data->>? IS NOT NULL) ORDER BY " +
		"data->>? ASC limit 1;", language, language)
	if err != nil {
		return "", errors.Trace(err)
	}

	var updateId string
	err = row.Scan(&updateId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return "", nil
		} else {
			return "", errors.Trace(err)
		}
	}

	return updateId, nil
}


func GetNewUpdatesWrappers(id int64) ([]*UpdateWrapper, error) {
	var wrappers []*UpdateWrapper

	res := getFromDB("wrappers", db.Cond{"id >" : id}).OrderBy("-id")

	exists, err := res.Exists()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if exists {
		res.All(&wrappers)
		return wrappers, nil
	} else {
		return nil, nil
	}
}


func GetLastUpdateWrappers(limit int) (updates []*UpdateWrapper, err error) {
	err = errors.Trace(getFromDB("wrappers", db.Cond{}).OrderBy("-id").
		Limit(limit).All(&updates))
	return
}


func GetUpdateWrappers(id int64, limit int) (updates []*UpdateWrapper, err error) {
	var wrappers []*UpdateWrapper

	res := GetInstance().db.Collection("wrappers").Find(db.Cond{"id <" : id}).OrderBy("-id").Limit(limit)

	exists, err := res.Exists()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if exists {
		res.All(&wrappers)
		return wrappers, nil
	} else {
		return nil, nil
	}
}


func (wrapper *UpdateWrapper) InsertToDB() error {
	_, err := GetInstance().db.InsertInto("wrappers").Values(wrapper).Exec()
	if err != nil {
		return errors.Trace(err)
	}

	rows, err := GetInstance().db.Select("id").From("wrappers").OrderBy("-id").Limit(1).Query()
	if err != nil {
		return errors.Trace(err)
	}

	if rows.Next() {
		rows.Scan(&wrapper.Id)
	}

	return nil
}


func (wrapper *UpdateWrapper) SaveToDB() error {
	return errors.Trace(updateInDB("wrappers", db.Cond{"id": wrapper.Id}, wrapper))
}


func (wrapper *UpdateWrapper) GetAvailableLang() []string {
	raw := wrapper.Data.V.(map[string]interface{})
	res := make([]string, len(raw))
	i := 0

	for key := range raw {
		res[i] = key
		i++
	}

	return res
}


func (wrapper *UpdateWrapper) HasLang(lang string) bool {
	raw := wrapper.Data.V.(map[string]interface{})

	for key := range raw {
		if key == lang {
			return true
		}
	}

	return false
}


func (wrapper *UpdateWrapper) GetUpdateId(lang string) int64 {
	raw := wrapper.Data.V.(map[string]interface{})

	for key, val := range raw {
		if key == lang {
			return int64(val.(float64))
		}
	}

	return -1
}


func (wrapper *UpdateWrapper) AddLastUpdates(lang string, updateId int64) {
	raw := wrapper.Data.V.(map[string]interface{})
	raw[lang] = updateId
}