package image_service

import (
	"bytes"
	"image"
	"image/jpeg"
	"github.com/nfnt/resize"
	"github.com/juju/errors"
	"net/http"

)


func DownloadImage(url string) (image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer response.Body.Close()

	img, _, err := image.Decode(response.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return img, nil
}

// as arguments
func ResizeImage(img image.Image, width, height uint)([]byte, error){
	imageResized := resize.Thumbnail (width, height, img, resize.MitchellNetravali)
	buf := new(bytes.Buffer)

	err := jpeg.Encode(buf, imageResized, nil)
	if err != nil {
		return nil,errors.Trace(err)
	}

	return buf.Bytes(), nil
}