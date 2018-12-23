package main

import (
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/engine"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/services/firebase_notifications"
)

func main() {
	go firebase_notifications.Start()

	engine.Start()
}
