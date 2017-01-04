package asset

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"

	log "github.com/gpitfield/relog"
	"github.com/nfnt/resize"
	"github.com/spf13/viper"
)

type SrcImage struct {
	Name   string
	Bounds image.Rectangle
	Suffix string
	XVal   string
	WVal   string
}

// RespImages return a slice of the SrcImages the given image should be resized to
func RespImages(inPath string, outDir string, baseName string, extension string, metaOnly bool) (srcSet []SrcImage) {
	in, err := os.Open(inPath)
	if err != nil {
		log.Error(err)
	}
	defer in.Close()
	img, _, err := image.Decode(in)
	if err != nil {
		log.Error(err)
	}
	bounds := img.Bounds()
	srcSet = []SrcImage{SrcImage{
		Bounds: bounds,
		Suffix: "",
		XVal:   "3x", // TBD how we do this
		WVal:   fmt.Sprintf("%dw", bounds.Max.X),
		Name:   baseName + extension,
	}}
	suffix := 1
	if !metaOnly {
		inCopy, err := os.Open(inPath)
		if err != nil {
			log.Error(err)
		}
		defer inCopy.Close()
		outFile, err := os.Create(outDir + "/" + baseName + extension)
		if err != nil {
			log.Error(err)
		}
		_, err = io.Copy(outFile, inCopy) // use an exact copy so the hash doesn't mutate
		if err != nil {
			log.Error(err)
		}
		outFile.Close()
	}

	for bounds.Max.X > 100 {
		suffix *= 2
		bounds.Max.X /= 2
		bounds.Max.Y /= 2
		src := SrcImage{
			Bounds: bounds,
			Suffix: fmt.Sprintf("_%d", suffix),
			XVal:   "tbd",
			WVal:   fmt.Sprintf("%dw", bounds.Max.X),
		}
		src.Name = baseName + src.Suffix + extension
		srcSet = append(srcSet, src)
		if !metaOnly {
			newImage := resize.Resize(uint(bounds.Max.X), uint(bounds.Max.Y), img, resize.Lanczos3)
			outFile, err := os.Create(outDir + "/" + src.Name)
			if err != nil {
				log.Error(err)
			}
			err = jpeg.Encode(outFile, newImage, &jpeg.Options{Quality: viper.GetInt("jpeg-quality")})
			if err != nil {
				log.Error(err)
			}
			outFile.Close()
		}
	}
	return
}
