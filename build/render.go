package build

import (
	"bytes"
	"html/template"
	"os"
	"strings"

	"github.com/GeertJohan/go.rice"
	log "github.com/gpitfield/relog"
	"github.com/spf13/viper"
)

func loadTemplates(list ...string) *template.Template {
	templateBox, err := rice.FindBox("templates")
	if err != nil {
		log.Fatal(err)
	}
	templates := template.New("")
	for _, x := range list {
		templateString, err := templateBox.String(x)
		if err != nil {
			log.Fatal(err)
		}
		_, err = templates.New(x).Parse(templateString)
		if err != nil {
			log.Fatal(err)
		}
	}

	funcMap := template.FuncMap{
		"title": func(a string) string { return strings.Title(a) },
		"lower": strings.ToLower,
	}
	templates.Funcs(funcMap)

	return templates
}

type NavInfo struct {
	Name string
	Link string
}

func renderAbout(collections []NavInfo) *bytes.Buffer {
	os.Mkdir(PubSiteDir+"/"+AboutDir, os.ModeDir|os.ModePerm)
	about := make(map[string]interface{})
	about["Title"] = viper.GetString("about-title")
	about["About"] = true
	about["Headline"] = viper.GetString("about-headline")
	aboutText := viper.GetStringSlice("about-text")
	aboutHtml := []template.HTML{}
	for _, el := range aboutText {
		aboutHtml = append(aboutHtml, template.HTML(el))
	}
	about["Text"] = aboutHtml
	about["Image"] = viper.GetString("about-image")
	about["Collections"] = collections
	buf := new(bytes.Buffer)
	templates.ExecuteTemplate(buf, "about.html", about)
	return buf

}

func renderDetail(collectionName string, collections []NavInfo, info PrintInfo, collectionInfo []PrintInfo) *bytes.Buffer {
	details := make(map[string]interface{})
	details["Collection"] = collectionName
	details["Collections"] = collections
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
	templates.ExecuteTemplate(buf, "detail.html", details)
	return buf
}

func renderGallery(collectionName string, collections []NavInfo, images []PrintInfo, collIndex int) *bytes.Buffer {
	root := viper.GetString("root-collection")
	gallery := make(map[string]interface{})
	gallery["Collection"] = collectionName
	gallery["Collections"] = collections
	gallery["Images"] = images
	gallery["Next"] = ""
	gallery["Previous"] = ""
	gallery["Gallery"] = true
	if len(collections) > 1 {
		if collIndex > 0 {
			if collections[collIndex-1].Name != root {
				gallery["Previous"] = "/" + collections[collIndex-1].Name
			}
		} else {
			if collections[len(collections)-1].Name != root {
				gallery["Previous"] = "/" + collections[len(collections)-1].Name
			}
		}
		if collIndex < len(collections)-1 {
			if collections[collIndex+1].Name != root {
				gallery["Next"] = "/" + collections[collIndex+1].Name
			}
		} else if collections[0].Name != root {
			gallery["Next"] = "/" + collections[0].Name
		}
	} else if collectionName != root {
		gallery["Next"] = "/" + collectionName
		gallery["Previous"] = "/" + collectionName
	}
	buf := new(bytes.Buffer)
	templates.ExecuteTemplate(buf, "gallery.html", gallery)
	return buf
}
