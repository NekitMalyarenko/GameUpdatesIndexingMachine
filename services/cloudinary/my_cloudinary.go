package cloudinary

import (
	"net/http"
	"github.com/juju/errors"
	"io/ioutil"

	"golang.org/x/net/context"
	"io"
	"github.com/NekitMalyarenko/GameUpdatesIndexingMachine/services/cloudinary/cloudinary_root"
	"strconv"
	"bytes"
	"os"
)


func Save(gameId int64, updateId, lang, updateHTML string) (url string, err error) {
	ctx := context.Background()
	ctx = cloudinary_root.NewContext(ctx, os.Getenv("cloudinary"))

	path := "game_updates/" + strconv.FormatInt(gameId, 10) + "/" + updateId + "/" + lang + ".raw"

	return "https://res.cloudinary.com/dbogdiydy/raw/upload/" + path, errors.Trace(cloudinary_root.UploadStaticRaw(ctx,
		path, bytes.NewBuffer([]byte(updateHTML))))
}


func SaveImage(image []byte,gameId int64,updateId string,imageId int64)(url string,err error){
	ctx := context.Background()
	ctx = cloudinary_root.NewContext(ctx,os.Getenv("cloudinary"))

	path := "game_updates/" + strconv.FormatInt(gameId, 10) + "/" + updateId +
		"/" + strconv.FormatInt(imageId, 10) + ".jpg"

	// for html source tag in image
	//path := "/get-image/" + strconv.FormatInt(gameId, 10) + "/" + updateId +
	//	"/" + strconv.FormatInt(imageId,10) +".jpg"

	return "https://res.cloudinary.com/dbogdiydy/raw/upload/" + path, errors.Trace(cloudinary_root.UploadStaticRaw(ctx,path,bytes.NewBuffer(image)))
}


func Test(data io.Reader) error {
	ctx := context.Background()
	ctx = cloudinary_root.NewContext(ctx, "cloudinary://245738261838881:lSLutX6LmWZKc4hfYPENoMUgCGg@dbogdiydy")

	return cloudinary_root.UploadStaticRaw(ctx, "game_updates/0/2018.08.30/test.html", data)
}


func Get(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", errors.Trace(err)
	}
	defer response.Body.Close()

	raw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}