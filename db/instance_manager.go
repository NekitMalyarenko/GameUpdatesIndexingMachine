package db

import (
	"upper.io/db.v3/postgresql"
	"sync"
	"log"
	"upper.io/db.v3/lib/sqlbuilder"
	"os"
	"github.com/juju/errors"
)

type dbService struct {
	db sqlbuilder.Database
}


var (
	instance *dbService

	mu sync.Mutex
)


func GetInstance() *dbService {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		settings, err := postgresql.ParseURL(os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Println(errors.Details(errors.Trace(err)))
			return nil
		}

		settings.Options = map[string]string{"sslmode" : "require"}

		log.Println("-----New DB instance-----")
		//log.Println(settings)
		newInstance, err := postgresql.Open(settings)
		if err != nil {
			log.Fatal(err)
		}

		instance = &dbService{
			db: newInstance,
		}
	}

	return instance
}


func CloseConnection() {
	if instance != nil {
		instance.db.Close()
	}
}