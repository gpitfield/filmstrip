// Package asset cuts and manages image assets based on the source folder, build settings,
// and deployed state
package asset

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"strconv"
	"strings"

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

func Hash(asset []byte) string {
	hash := md5.New()
	hash.Write(asset)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Given an image filename, decode its "real" name, order position, whether it is a cover image, and if it's untitled
func FileInfo(filename string) (name string, order int, cover bool, untitled bool) {
	name = filename
	if strings.HasPrefix(filename, "*") {
		cover = true
		filename = "_" + filename[1:]
	}
	if strings.HasPrefix(filename, "_") {
		parts := strings.SplitN(strings.TrimLeft(filename, "_"), "_", 2)
		if len(parts) == 2 {
			ord, err := strconv.Atoi(parts[0])
			if err != nil {
				log.Error(err)
			}
			name = parts[1]
			order = ord
		}
	}
	if viper.GetBool("auto-untitle") && (strings.Contains(name, "dsc") || strings.Contains(name, "DSC")) {
		untitled = true
	}
	return
}

// RespImages return a slice of the SrcImages the given image should be resized to
func RespImages(inPath string, outDir string, baseName string, extension string, metaOnly bool) (srcSet []SrcImage) {
	var (
		bounds image.Rectangle
		img    image.Image
		err    error
	)
	in, err := os.Open(inPath)
	if err != nil {
		log.Error(err)
	}
	defer in.Close()
	if metaOnly {
		cfg, _, err := image.DecodeConfig(in)
		if err != nil {
			log.Error(err)
		}
		bounds = image.Rect(0, 0, cfg.Width, cfg.Height)
	} else {
		img, _, err = image.Decode(in)
		if err != nil {
			log.Error(err)
		}
		bounds = img.Bounds()
	}
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
