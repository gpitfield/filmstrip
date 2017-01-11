// Package site handles site structure and local sync
package site

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/gpitfield/filmstrip/asset"
	log "github.com/gpitfield/relog"
	"github.com/spf13/viper"
)

const (
	PubSiteDir    = "public"
	SiteDir       = "site"
	CSSStylesDir  = "css"
	JavaScriptDir = "js"
	ImagesDir     = "images"
	AboutDir      = "about"
	FilmstripCSS  = "filmstrip.css"
)

var Templates *template.Template

func init() {
	Templates = loadTemplates("detail.html", "bootstrap.html", "nav.html", "gallery.html", "cover.html", "bottom-nav.html", "about.html", "filmstrip.css")
}

// Escape returns the lowercased, dash-spaced, and escaped version of the input
func Escape(in string) string {
	return url.QueryEscape(LowerDash(in))
}

func LowerDash(in string) string {
	return strings.Replace(strings.ToLower(in), " ", "-", -1)
}

// Flush removes any invalid/remnant folders from the local public site folder
func Flush(prefix, name string) {
	var (
		err                  error
		sourceLocation       = viper.GetString("source-dir")
		sourceDirs, siteDirs []os.FileInfo
		fullPath             = "/"
	)
	protectedSubDirs := []string{ImagesDir, AboutDir, CSSStylesDir, JavaScriptDir}
	if prefix != "" {
		fullPath = "/" + prefix + "/" + name
	} else if name != "" {
		fullPath = "/" + name
	}
	sourceDirs, err = ioutil.ReadDir(sourceLocation + fullPath)
	checkErr(err)
	sourceDirs, err = ioutil.ReadDir(PubSiteDir + LowerDash(fullPath))
	checkErr(err)

	for _, dir := range sourceDirs {
		if dir.IsDir() {
			protectedSubDirs = append(protectedSubDirs, dir.Name())
		}
	}

	for _, dir := range siteDirs {
		if !dir.IsDir() {
			continue
		}
		remove := true
		for _, sub := range protectedSubDirs {
			if sub == dir.Name() {
				remove = false
				break
			}
		}
		if remove {
			err := os.RemoveAll(PubSiteDir + LowerDash(fullPath) + "/" + dir.Name())
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func FlushInvalid(dir string, validFiles map[string]bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Error(err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !validFiles[f.Name()] {
			remove := true
			for key, _ := range validFiles {
				if strings.HasPrefix(f.Name(), key) {
					remove = false
					break
				}
			}
			if remove {
				os.Remove(dir + "/" + f.Name()) // delete invalid file
			}
		}
	}
}

// Scaffold copies CSS, JS and similar scaffolding to the local public site directory
// TODO: these assets should have a hash set as part of their filename to enable more effective caching/busting
func Scaffold() {
	mode := os.ModeDir | os.ModePerm
	checkErr(os.Mkdir(PubSiteDir, mode))
	checkErr(os.Mkdir(PubSiteDir+"/"+CSSStylesDir, mode))
	checkErr(os.Mkdir(PubSiteDir+"/"+JavaScriptDir, mode))
	checkErr(os.Mkdir(PubSiteDir+"/"+AboutDir, mode))

	for _, dir := range []string{JavaScriptDir} {
		copyScaffoldDir(dir)
	}
	css := make(map[string]interface{})
	css["CoverCols"] = viper.GetString("cover-columns")
	css["GalleryCols"] = viper.GetString("gallery-columns")
	var page bytes.Buffer
	Templates.ExecuteTemplate(&page, FilmstripCSS, css)
	checkErr(ioutil.WriteFile(PubSiteDir+"/"+CSSStylesDir+"/"+FilmstripCSS, page.Bytes(), 0644))
}

func copyScaffoldDir(dir string) {
	files, err := ioutil.ReadDir(SiteDir + "/" + dir)
	if err != nil {
		log.Error(err)
	}
	for _, file := range files {
		var (
			inPath  = dir + "/" + file.Name()
			outPath = PubSiteDir + "/" + dir + "/" + file.Name()
		)

		inBytes, _ := ioutil.ReadFile(inPath)
		outBytes, err := ioutil.ReadFile(outPath)
		if err != nil {
			log.Error(err)
		} else if asset.Hash(inBytes) == asset.Hash(outBytes) {
			continue
		}
		out, err := os.Create(outPath)
		if err != nil {
			log.Error(err)
		}
		defer out.Close()
		_, err = out.Write(inBytes)
		if err != nil {
			log.Error(err)
		}
	}
}

func loadTemplates(list ...string) *template.Template {
	templateBox, err := rice.FindBox("templates")
	if err != nil {
		log.Fatal(err)
	}
	funcMap := template.FuncMap{
		"title":  func(a string) string { return strings.Title(a) },
		"lower":  strings.ToLower,
		"escape": url.QueryEscape,
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	templates := template.New("").Funcs(funcMap)
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

	return templates
}

func checkErr(err error) {
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
}
