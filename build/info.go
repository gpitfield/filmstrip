package build

import (
	"fmt"
	// "image"
	_ "image/jpeg"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gpitfield/filmstrip/asset"
	"github.com/gpitfield/filmstrip/site"
	log "github.com/gpitfield/relog"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/viper"
)

type PrintInfo struct {
	Filename     string
	Title        string
	Untitled     bool
	FileURL      string
	RelURL       string // dash-spaced, lowercased, and escaped
	AbsURL       string // dash-spaced, lowercased, and escaped
	IncludesExif bool
	Order        int
	Description  string
	Date         time.Time
	DateString   string
	CameraInfo   string
	Copyright    string
	Cover        bool
	SrcImages    []asset.SrcImage
}

type NavInfo struct {
	Name string
	Link string
}

type PageInfo struct {
	Title     string
	URLTitle  string // dash-spaced, lowercased, and escaped version of Title
	SubHed    string
	Details   string
	Copyright string
	Bug       string
	PrintInfo
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

func getInfo(filename string, r io.Reader) (info PrintInfo) {
	info.Filename = filename
	title, order, cover, untitled := asset.FileInfo(filename)
	info.Title = stripExtension(title)
	info.Untitled = untitled
	info.RelURL = site.Escape(info.Title)
	info.FileURL = site.Escape(info.Filename)
	info.Order = order
	info.Cover = cover

	if r == nil {
		return
	}
	x, err := exif.Decode(r)
	if err != nil {
		log.Error(filename)
		log.Fatal(err)
	}
	info.IncludesExif = true
	zoom := false
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
			info.DateString = info.Date.Format("January, 2006")
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
		if strings.Contains(tag.String(), "-") {
			zoom = true
		}
	}
	if zoom {
		if tag, err := x.Get(exif.FocalLengthIn35mmFilm); err == nil {
			if info.CameraInfo != "" {
				info.CameraInfo += " | "
			}
			info.CameraInfo += tag.String() + "mm"
		}
	}
	return
}

func stripExtension(in string) string {
	split := strings.Split(in, ".")
	if len(split) > 1 {
		return strings.Join(split[0:len(split)-1], ".")
	}
	return in
}

func extension(in string) string {
	split := strings.Split(in, ".")
	if len(split) > 1 {
		return "." + split[len(split)-1]
	}
	return in
}
