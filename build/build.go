package build

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/gpitfield/relog"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/spf13/viper"
)

const (
	PubSiteDir   = "site"
	CSSStylesDir = "css"
)

var (
	sourceLocation string
	templates      *template.Template
)

type PrintInfo struct {
	Filename    string
	Title       string
	Order       int
	Description string
	Date        time.Time
	DateString  string
	CameraInfo  string
	Copyright   string
}

type Ordered []PrintInfo

func (o Ordered) Len() int      { return len(o) }
func (o Ordered) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o Ordered) Less(i, j int) bool {
	if o[j].Order == 0 && o[j].Order == 0 {
		return o[i].Date.Before(o[j].Date)
	}
	if o[j].Order == 0 {
		return true
	}
	if o[i].Order == 0 {
		return false
	}
	return o[i].Order < o[j].Order
}

// Build the site from the provided content
func Build() {
	start := time.Now()
	flushPrevious()                               // flush all previous data from PubSiteDir
	collections := getCollections(sourceLocation) // get the directories within the sourceLocation
	var navs = make([]NavInfo, len(collections))
	root := viper.GetString("root-collection")
	for i, _ := range collections {
		navs[i] = NavInfo{Name: collections[i]}
		if collections[i] == root {
			navs[i].Link = "/index.html"
		} else {
			navs[i].Link = "/" + collections[i] + "/index.html"
		}
	}
	templates = loadTemplates("detail.html", "bootstrap.html", "nav.html", "gallery.html", "bottom-nav.html")
	for i, _ := range collections {
		buildCollection(i, collections, navs, collections[i] == root)
	}
	cssFiles, err := ioutil.ReadDir(CSSStylesDir)
	if err != nil {
		log.Error(err)
	}
	for _, file := range cssFiles {
		in, err := os.Open(CSSStylesDir + "/" + file.Name())
		if err != nil {
			log.Error(err)
		}
		out, err := os.Create(PubSiteDir + "/" + CSSStylesDir + "/" + file.Name())
		if err != nil {
			log.Error(err)
		}
		_, err = io.Copy(out, in)
		if err != nil {
			log.Error(err)
		}
		in.Close()
		out.Close()
	}
	log.Infof("built in %v", time.Since(start))
}

func buildCollection(index int, colls []string, navs []NavInfo, root bool) {
	collectionDirs(colls[index], root)
	var imageInfo []PrintInfo
	images, err := ioutil.ReadDir(sourceLocation + "/" + colls[index])
	if err != nil {
		log.Error(err)
	}
	for _, image := range images {
		if image.IsDir() {
			log.Warnf("Sub-collections not supported (%s/%s)", colls[index], image.Name())
			continue
		}
		info := getInfo(sourceLocation+"/"+colls[index], image.Name())
		imageInfo = append(imageInfo, info)
		copyImage(colls[index], info, root)
	}
	sort.Sort(Ordered(imageInfo))
	var page, gallery *bytes.Buffer
	for _, info := range imageInfo {
		if root && viper.GetString("home-title") != "" {
			page = renderDetail(viper.GetString("home-title"), navs, info, imageInfo)
		} else {
			page = renderDetail(colls[index], navs, info, imageInfo)
		}
		if root {
			err = ioutil.WriteFile(PubSiteDir+"/"+info.Title+".html", page.Bytes(), 0644)
		} else {
			err = ioutil.WriteFile(PubSiteDir+"/"+colls[index]+"/"+info.Title+".html", page.Bytes(), 0644)
		}
		if err != nil {
			log.Error(err)
		}
	}
	if root && viper.GetString("home-title") != "" {
		gallery = renderGallery(viper.GetString("home-title"), navs, imageInfo, index)
	} else {
		gallery = renderGallery(colls[index], navs, imageInfo, index)
	}
	if root {
		err = ioutil.WriteFile(PubSiteDir+"/index.html", gallery.Bytes(), 0644)
	} else {
		err = ioutil.WriteFile(PubSiteDir+"/"+colls[index]+"/index.html", gallery.Bytes(), 0644)
	}
	if err != nil {
		log.Error(err)
	}
}

func copyImage(collection string, info PrintInfo, root bool) {
	var (
		in, out *os.File
		err     error
	)

	in, err = os.Open(sourceLocation + "/" + collection + "/" + info.Filename)
	if err != nil {
		log.Error(err)
	}
	if root {
		out, err = os.Create(PubSiteDir + "/images/" + info.Filename)
	} else {
		out, err = os.Create(PubSiteDir + "/" + collection + "/images/" + info.Filename)
	}
	if err != nil {
		log.Error(err)
	}
	_, err = io.Copy(out, in)
	if err != nil {
		log.Error(err)
	}
	in.Close()
	out.Close()
}

func collectionDirs(collection string, root bool) {
	if root {
		err := os.Mkdir(PubSiteDir+"/images", os.ModeDir|os.ModePerm)
		if err != nil {
			log.Error(err)
		}
	} else {
		err := os.Mkdir(PubSiteDir+"/"+collection, os.ModeDir|os.ModePerm)
		if err != nil {
			log.Error(err)
		}
		err = os.Mkdir(PubSiteDir+"/"+collection+"/images", os.ModeDir|os.ModePerm)
		if err != nil {
			log.Error(err)
		}
	}
}

func stripExtension(in string) string {
	split := strings.Split(in, ".")
	if len(split) > 1 {
		return strings.Join(split[0:len(split)-1], ".")
	}
	return in
}

func getInfo(path string, filename string) (info PrintInfo) {
	f, err := os.Open(path + "/" + filename)
	if err != nil {
		log.Error(err)
		return
	}
	x, err := exif.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	info.Filename = filename
	info.Title = stripExtension(filename)
	if strings.HasPrefix(filename, "_") {
		parts := strings.SplitN(strings.TrimLeft(filename, "_"), "_", 2)
		if len(parts) == 2 {
			pos, err := strconv.Atoi(parts[0])
			if err != nil {
				log.Error(err)
			} else {
				info.Order = pos
				info.Title = stripExtension(parts[1])
			}
		}
	}
	if tag, err := x.Get(exif.Copyright); err == nil && tag.String() != "" {
		info.Copyright = strings.Trim(tag.String(), "\"")
	} else {
		info.Copyright = viper.GetString("copyright")
	}

	if tag, err := x.Get(exif.ImageDescription); err == nil && tag.String() != "" {
		info.Description = strings.Trim(tag.String(), "\"")
	}

	if tag, err := x.Get(exif.DateTimeOriginal); err == nil && tag.String() != "" {
		date, err := time.Parse("\"2006:01:02 15:04:05\"", tag.String())
		if err != nil {
			log.Error(err)
		} else {
			info.Date = date
			info.DateString = info.Date.Format("January 2, 2006")
		}
	}

	if fstop, err := x.Get(exif.FNumber); err == nil {
		f := strings.Split(strings.Trim(fstop.String(), "\""), "/")
		var fNum = make([]int, 2)
		fNum[0], err = strconv.Atoi(f[0])
		if err != nil {
			log.Error(err)
		}
		fNum[1], err = strconv.Atoi(f[1])
		if err != nil {
			log.Error(err)
		}
		if fNum[1] == 1 {
			info.CameraInfo = fmt.Sprintf("f/%d", fNum[0])
		} else {
			info.CameraInfo = fmt.Sprintf("f/%.1f", float64(fNum[0])/float64(fNum[1]))
		}
	} else {
		log.Error(err)
	}

	if tag, err := x.Get(exif.ExposureTime); err == nil {
		xTime := strings.Trim(tag.String(), "\"")
		if info.CameraInfo != "" {
			info.CameraInfo += " | "
		}
		info.CameraInfo += xTime + "s"
	}

	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if info.CameraInfo != "" {
			info.CameraInfo += " | "
		}
		info.CameraInfo += "ISO " + tag.String()
	}

	if tag, err := x.Get(exif.FocalLengthIn35mmFilm); err == nil {
		if info.CameraInfo != "" {
			info.CameraInfo += " | "
		}
		info.CameraInfo += tag.String() + "mm"
	}

	if tag, err := x.Get(exif.Model); err == nil {
		if info.CameraInfo != "" {
			info.CameraInfo += " | "
		}
		info.CameraInfo += strings.Trim(tag.String(), "\"")
	}

	if tag, err := x.Get(exif.LensModel); err == nil {
		if info.CameraInfo != "" {
			info.CameraInfo += " | "
		}
		info.CameraInfo += strings.Trim(tag.String(), "\"")
	}
	return
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

func flushPrevious() {
	subDirs, err := ioutil.ReadDir(PubSiteDir)
	if err != nil {
		log.Error(err)
	}
	for _, dir := range subDirs {
		err := os.RemoveAll(PubSiteDir + "/" + dir.Name())
		if err != nil {
			log.Error(err)
		}
	}
	err = os.Mkdir(PubSiteDir, os.ModeDir|os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
	err = os.Mkdir(PubSiteDir+"/"+CSSStylesDir, os.ModeDir|os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
}

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
