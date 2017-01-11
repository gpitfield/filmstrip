package build

import (
	"bytes"
	"html/template"

	// log "github.com/gpitfield/relog"
	"github.com/gpitfield/filmstrip/site"
	"github.com/spf13/viper"
)

func renderAbout() *bytes.Buffer {
	about := make(map[string]interface{})
	about["Title"] = viper.GetString("site-title")
	about["About"] = true
	about["Headline"] = viper.GetString("about-headline")
	aboutText := viper.GetStringSlice("about-text")
	aboutHtml := []template.HTML{}
	for _, el := range aboutText {
		aboutHtml = append(aboutHtml, template.HTML(el))
	}
	about["Text"] = aboutHtml
	about["Image"] = viper.GetString("about-image")
	buf := new(bytes.Buffer)
	site.Templates.ExecuteTemplate(buf, "about.html", about)
	return buf

}

func renderDetail(collectionName string, info PrintInfo, collectionInfo []PrintInfo) *bytes.Buffer {
	details := make(map[string]interface{})
	details["Title"] = viper.GetString("site-title")
	details["Collection"] = collectionName
	if len(collectionInfo) > 1 {
		var matchIndex = 0
		for i, _ := range collectionInfo {
			if collectionInfo[i].Filename == info.Filename {
				matchIndex = i
				break
			}
		}
		if matchIndex > 0 {
			details["Previous"] = collectionInfo[matchIndex-1]
		} else {
			details["Previous"] = collectionInfo[len(collectionInfo)-1]
		}
		if matchIndex < len(collectionInfo)-1 {
			details["Next"] = collectionInfo[matchIndex+1]
		} else {
			details["Next"] = collectionInfo[0]
		}
	} else {
		details["Next"] = 0
		details["Previous"] = 0
	}
	details["Image"] = info
	buf := new(bytes.Buffer)
	site.Templates.ExecuteTemplate(buf, "detail.html", details)
	return buf
}

func renderGallery(collectionName string, images []PrintInfo, cover bool) *bytes.Buffer {
	gallery := make(map[string]interface{})
	gallery["Gallery"] = true
	if collectionName == "" {
		gallery["Collection"] = viper.GetString("site-title")
		gallery["Home"] = true
	} else {
		gallery["Collection"] = collectionName
	}
	gallery["Title"] = viper.GetString("site-title")
	gallery["Copyright"] = viper.GetString("copyright")
	gallery["Images"] = images
	buf := new(bytes.Buffer)
	if cover {
		site.Templates.ExecuteTemplate(buf, "cover.html", gallery)
	} else {
		site.Templates.ExecuteTemplate(buf, "gallery.html", gallery)
	}
	return buf
}
