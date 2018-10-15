package main

import (
	"os"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/engine"
)

func main() {
	os.Setenv("cloudinary", "cloudinary://245738261838881:lSLutX6LmWZKc4hfYPENoMUgCGg@dbogdiydy")

	os.Setenv("DATABASE_URL", "postgres://vdiyfvhzesxlfv:c38fb964f27179ffbfcc42d654b1228d6f10ad1b7d661d3895a5" +
		"3659de54c3de@ec2-79-125-12-48.eu-west-1.compute.amazonaws.com:5432/d8t0f9lphlqkom")

	engine.Start()
}
