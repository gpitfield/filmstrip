package build

import (
	"io/ioutil"
	"os"

	"github.com/gpitfield/filmstrip/asset"
	log "github.com/gpitfield/relog"
)

// Scaffold copies CSS, JS and similar scaffolding to the site directory
// TODO: these assets should have a hash set as part of their filename to enable more effective caching
func Scaffold() {
	for _, dir := range []string{CSSStylesDir, JavaScriptDir} {
		files, err := ioutil.ReadDir(dir)
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
}
