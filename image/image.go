package image

import (
	"bytes"
	"image/jpeg"
	"io/ioutil"
	"log"

	"github.com/disintegration/imaging"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

func TransformToWebP(b []byte, width int, height int) ([]byte, error) {

	img, err := jpeg.Decode(bytes.NewBuffer(b))
	if err != nil {
		log.Fatalln(err)
	}
	img = imaging.Resize(img, width, height, imaging.Lanczos)

	buf := bytes.NewBuffer([]byte(""))
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
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
