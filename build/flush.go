package build

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/gpitfield/filmstrip/asset"
	log "github.com/gpitfield/relog"
)

func flushPrevious(collections []string) {
	asset.FlushDir(PubSiteDir, collections)
	err := os.Mkdir(PubSiteDir, os.ModeDir|os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
	err = os.Mkdir(PubSiteDir+"/"+CSSStylesDir, os.ModeDir|os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
	err = os.Mkdir(PubSiteDir+"/"+JavaScriptDir, os.ModeDir|os.ModePerm)
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
}

func flushInvalid(dir string, validFiles map[string]bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Error(err)
	}
	for _, f := range files {
		if f.IsDir() {
			flushInvalid(dir+"/"+f.Name(), validFiles)
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
