package db

import (
	"upper.io/db.v3"
)

var (

	getFromDB = func(table string, cond db.Cond) (db.Result)  {
		return GetInstance().db.Collection(table).
			Find(cond)
	}



	insertToDB = func(table string, val interface{}) error {
		_, err := GetInstance().db.InsertInto(table).Values(val).
			Exec()
		return err
	}

	updateInDB = func(table string, cond db.Cond, val interface{}) error {
		res := GetInstance().db.Collection(table).Find(cond)
		return res.Update(val)
	}

	deleteFromDB = func(table string, cond db.Cond) error {
		_, err := GetInstance().db.DeleteFrom(table).
			Where(cond).Exec()
		return err
	}
)