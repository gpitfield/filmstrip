package build

import (
	"bytes"
	"fmt"
	"html/template"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gpitfield/filmstrip/asset"
	log "github.com/gpitfield/relog"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/spf13/viper"
)

const (
	PubSiteDir    = "site"
	CSSStylesDir  = "css"
	JavaScriptDir = "js"
	ImagesDir     = "images"
	AboutDir      = "about"
)

var (
	sourceLocation string
	templates      *template.Template
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // config file in working directory
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	sourceLocation = viper.GetString("source-dir") // build needs a source folder
	exif.RegisterParsers(mknote.All...)
}

// Build the site from the provided image content where changed. If force is true, regenerate the site regardless of changes,
// typically if the HTML has changed. Force will not copy over unchanged images, it just regenerates associated HTML.
func Build(force bool) {
	start := time.Now()
	templates = loadTemplates("detail.html", "bootstrap.html", "nav.html", "gallery.html", "bottom-nav.html", "about.html")
	Scaffold()

	collections := getCollections(sourceLocation)                                        // get the directories within the sourceLocation
	flushPrevious(append(collections, ImagesDir, AboutDir, CSSStylesDir, JavaScriptDir)) // flush all expired/deleted data from PubSiteDir
	// root := viper.GetString("root-collection")
	var navs = make([]NavInfo, len(collections))
	for i, _ := range collections {
		navs[i] = NavInfo{
			Name: collections[i],
			Link: "/" + collections[i] + "/index.html",
		}
	}
	for i, _ := range collections {
		// buildCollection(i, collections, navs, collections[i] == root)
		buildCollection(i, collections, navs, force)
	}

	buildAbout(navs)
	log.Infof("built in %v", time.Since(start))
}

func buildAbout(navs []NavInfo) {
	about := renderAbout(navs)
	err := ioutil.WriteFile(PubSiteDir+"/"+AboutDir+"/index.html", about.Bytes(), 0644)
	if err != nil {
		log.Error(err)
	}

	inCopy, err := os.Open(viper.GetString("about-image"))
	if err != nil {
		log.Error(err)
	}
	defer inCopy.Close()
	outFile, err := os.Create(PubSiteDir + "/" + AboutDir + "/" + "about.jpg")
	if err != nil {
		log.Error(err)
	}
	_, err = io.Copy(outFile, inCopy) // use an exact copy so the hash doesn't mutate
	if err != nil {
		log.Error(err)
	}
	outFile.Close()
}

func buildCollection(index int, colls []string, navs []NavInfo, force bool) {
	collectionDirs(colls[index])
	var (
		imageInfo     []PrintInfo
		validFiles    = map[string]bool{}
		inPath        = sourceLocation + "/" + colls[index]
		outPath       = PubSiteDir + "/" + colls[index]
		outImagesPath = outPath + "/" + ImagesDir
	)
	images, err := ioutil.ReadDir(inPath)
	if err != nil {
		log.Error(err)
	}

	// go through the images, and if they are changed vs the site folder, cut and copy them over
	// flush any images that are no longer in the collection
	// using the imageInfo slice, generate the HTML for each image as well as the gallery
	// flush any HTML for images no longer present

	for _, image := range images {
		if image.IsDir() {
			log.Warnf("Sub-collections not yet supported (%s/%s)", colls[index], image.Name())
			continue
		}
		// see if image has changed
		source, _ := ioutil.ReadFile(inPath + "/" + image.Name())
		inHash := asset.Hash(source)
		dest, _ := ioutil.ReadFile(outImagesPath + "/" + image.Name())
		outHash := asset.Hash(dest)
		if force || inHash != outHash {
			info := getInfo(image.Name(), io.Reader(bytes.NewReader(source)))
			srcs := copyImage(colls[index], info, inHash == outHash)
			info.SrcImages = srcs
			imageInfo = append(imageInfo, info)
			for _, src := range srcs {
				validFiles[src.Name] = true
			}
		} else {
			info := getInfo(image.Name(), nil) // get the basic info for gallery generation
			validFiles[stripExtension(image.Name())] = true
			imageInfo = append(imageInfo, info)
		}
	}

	sort.Sort(Ordered(imageInfo))
	var page, gallery *bytes.Buffer
	for _, info := range imageInfo {
		validFiles[stripExtension(info.Filename)] = true
		validFiles[info.Title] = true
		if info.IncludesExif {
			page = renderDetail(colls[index], navs, info, imageInfo)
			err = ioutil.WriteFile(PubSiteDir+"/"+colls[index]+"/"+info.Title+".html", page.Bytes(), 0644)
			validFiles[info.Title+".html"] = true
			if err != nil {
				log.Error(err)
			}
		}
		// if root && viper.GetString("home-title") != "" {
		// 	page = renderDetail(viper.GetString("home-title"), navs, info, imageInfo)
		// } else {
		// }
		// if root {
		// 	err = ioutil.WriteFile(PubSiteDir+"/"+info.Title+".html", page.Bytes(), 0644)
		// } else {
		// }
	}
	// if root && viper.GetString("home-title") != "" {
	// 	gallery = renderGallery(viper.GetString("home-title"), navs, imageInfo, index)
	// 	err = ioutil.WriteFile(PubSiteDir+"/index.html", gallery.Bytes(), 0644)
	// 	if err != nil {
	// 		log.Error(err)
	// 	}
	// }

	gallery = renderGallery(colls[index], navs, imageInfo, index)
	err = ioutil.WriteFile(PubSiteDir+"/"+colls[index]+"/index.html", gallery.Bytes(), 0644)
	validFiles["index.html"] = true
	if err != nil {
		log.Error(err)
	}
	flushInvalid(outPath, validFiles)
}

func copyImage(collection string, info PrintInfo, metaOnly bool) (srcSet []asset.SrcImage) {
	return asset.RespImages(sourceLocation+"/"+collection+"/"+info.Filename, PubSiteDir+"/"+collection+"/"+ImagesDir, stripExtension(info.Filename), extension(info.Filename), metaOnly)
}

func collectionDirs(collection string) {
	err := os.Mkdir(PubSiteDir+"/"+collection, os.ModeDir|os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
	err = os.Mkdir(PubSiteDir+"/"+collection+"/"+ImagesDir, os.ModeDir|os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
}

func getCollections(source string) (colls []string) {
	files, err := ioutil.ReadDir(source)
	if err != nil {
		log.Error(err)
	}
	for _, f := range files {
		if f.IsDir() {
			colls = append(colls, f.Name())
		}
	}
	return colls
}
