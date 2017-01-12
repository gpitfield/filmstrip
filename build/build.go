package build

import (
	"bytes"
	"fmt"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/gpitfield/filmstrip/asset"
	"github.com/gpitfield/filmstrip/site"
	log "github.com/gpitfield/relog"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/spf13/viper"
)

var sourceLocation string

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // config file in working directory
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal config file error: %s \n", err))
	}
	sourceLocation = viper.GetString("source-dir") // build needs a source folder
	exif.RegisterParsers(mknote.All...)
}

// Build the site from the provided image content where changed. If force is true, regenerate the site regardless of changes,
// typically if the HTML has changed. --force does not re-cut unchanged images, it just regenerates associated HTML.
func Build(force bool) {
	start := time.Now()
	site.Scaffold()
	buildCollection("", "", force)
	buildAbout()
	log.Infof("built in %v", time.Since(start))
}

// buildCollection recursively builds collections from a given name and path prefix
func buildCollection(prefix, name string, force bool) (coverInfo PrintInfo) {
	collectionDirs(prefix, name)
	var (
		imageInfo     []PrintInfo
		collName      string
		cover         bool
		order         int
		validFiles    = map[string]bool{"index.html": true}
		inPath        = prefix
		outPath       = site.LowerDash(prefix)
		outImagesPath string
	)

	if name != "" {
		collName, order, _, _ = asset.FileInfo(name)
		inPath += "/" + name
		outPath += "/" + site.LowerDash(collName)
	}
	site.Flush(prefix, name)
	outImagesPath = outPath + "/" + site.ImagesDir

	files, err := ioutil.ReadDir(sourceLocation + "/" + inPath)
	if err != nil {
		log.Error(err)
	}
	log.Infof("building collection %s/%s (%d files)", prefix, name, len(files))

	// recursively cut and copy changed/new images to local public site images
	for _, file := range files {
		if file.IsDir() {
			imageInfo = append(imageInfo, buildCollection(inPath, file.Name(), force))
			cover = true
			continue
		} else if collName == "" { // ignore any images at the topmost level
			continue
		} else if file.Name() == ".DS_Store" {
			continue
		}
		// see if file has changed
		source, _ := ioutil.ReadFile(sourceLocation + inPath + "/" + file.Name())
		inHash := asset.Hash(source)
		dest, _ := ioutil.ReadFile(site.PubSiteDir + outImagesPath + "/" + site.LowerDash(file.Name()))
		outHash := asset.Hash(dest)
		info := getInfo(file.Name(), io.Reader(bytes.NewReader(source)))
		validFiles[stripExtension(file.Name())] = true
		srcs := asset.RespImages(sourceLocation+inPath+"/"+file.Name(), site.PubSiteDir+outImagesPath, site.LowerDash(stripExtension(info.Filename)), extension(info.Filename), inHash == outHash)
		info.SrcImages = srcs
		info.AbsURL = outPath + "/" + info.RelURL
		imageInfo = append(imageInfo, info)
		for _, src := range srcs {
			validFiles[src.Name] = true
		}
	}

	sort.Sort(Ordered(imageInfo))
	var page, gallery *bytes.Buffer
	if len(imageInfo) > 0 {
		coverInfo = imageInfo[0] // default
	}
	// generate the HTML for each image as well as the gallery using imageInfo
	for _, info := range imageInfo {
		if info.Cover {
			coverInfo = info
		}
		validFiles[site.LowerDash(info.Title)+".html"] = true
		validFiles[stripExtension(info.Filename)] = true
		validFiles[info.Title] = true

		if !cover && info.IncludesExif {
			page = renderDetail(collName, info, imageInfo)
			err = ioutil.WriteFile(site.PubSiteDir+"/"+site.LowerDash(collName)+"/"+site.LowerDash(info.Title)+".html", page.Bytes(), 0644)
			if err != nil {
				log.Error(err)
			}
		}
	}
	coverInfo.Title = collName
	coverInfo.FileURL = site.LowerDash(collName)
	coverInfo.Order = order
	gallery = renderGallery(collName, imageInfo, cover)
	err = ioutil.WriteFile(site.PubSiteDir+"/"+site.LowerDash(collName)+"/index.html", gallery.Bytes(), 0644)
	if err != nil {
		log.Error(err)
	}

	site.FlushInvalid(site.PubSiteDir+outPath, validFiles) // flush any files associated with removed images
	return
}

func collectionDirs(parent, collection string) {
	mode := os.ModeDir | os.ModePerm
	if collection == "" {
		checkErr(os.Mkdir(site.LowerDash(site.PubSiteDir+"/"+parent+"/"+site.ImagesDir), mode))
	} else {
		checkErr(os.Mkdir(site.LowerDash(site.PubSiteDir+"/"+parent+"/"+collection), mode))
		checkErr(os.Mkdir(site.LowerDash(site.PubSiteDir+"/"+parent+"/"+collection)+"/"+site.ImagesDir, mode))
	}
}

// func buildAbout(navs []NavInfo) {
func buildAbout() {
	about := renderAbout()
	err := ioutil.WriteFile(site.PubSiteDir+"/"+site.AboutDir+"/index.html", about.Bytes(), 0644)
	if err != nil {
		log.Error(err)
	}

	inCopy, err := os.Open(viper.GetString("about-image"))
	if err != nil {
		log.Error(err)
	}
	defer inCopy.Close()
	outFile, err := os.Create(site.PubSiteDir + "/" + site.AboutDir + "/" + "about.jpg")
	if err != nil {
		log.Error(err)
	}
	_, err = io.Copy(outFile, inCopy) // use an exact copy so the hash doesn't mutate
	if err != nil {
		log.Error(err)
	}
	outFile.Close()
}
