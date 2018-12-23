package services

import (
	"bytes"
	"image"
	"image/jpeg"
	"github.com/nfnt/resize"
	"github.com/juju/errors"
)

// as arguments
func ResizeImage(img image.Image, width, height uint)([]byte, error){
	imageResized := resize.Resize(width ,height, img, resize.MitchellNetravali)
	buf := new(bytes.Buffer)

	err := jpeg.Encode(buf, imageResized, nil)
	if err != nil {
		return nil,errors.Trace(err)
	}

	return buf.Bytes(), nil
}