package image

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"

	"github.com/disintegration/imaging"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

func TransformToWebP(format string, b []byte, width int, height int) ([]byte, error) {

	lossless := false
	var img image.Image
	var err error
	if format == "image/jpeg" {
		img, err = jpeg.Decode(bytes.NewBuffer(b))
	} else if format == "image/png" {
		img, err = png.Decode(bytes.NewBuffer(b))
		lossless = true
	} else {
		return nil, errors.New("Cannot load image format: " + format)
	}
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	if width > 0 {
		img = imaging.Resize(img, width, height, imaging.Lanczos)
	}

	buf := bytes.NewBuffer([]byte(""))
	var options *encoder.Options
	if lossless {
		options, err = encoder.NewLosslessEncoderOptions(encoder.PresetDefault, 0)
	} else {
		options, err = encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	}
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	if err := webp.Encode(buf, img, options); err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return ioutil.ReadAll(buf)
}
