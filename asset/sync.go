// Package asset cuts image assets, and keeps assets in sync based on the source folder, build settings,
// and deployed state.
package asset

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/gpitfield/relog"
)

func Hash(asset []byte) string {
	hash := md5.New()
	hash.Write(asset)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// FlushDir flushes any invalid/remnant site folders
func FlushDir(parent string, allowedSubDirs []string) {
	subDirs, err := ioutil.ReadDir(parent)
	if err != nil {
		log.Error(err)
	}
	for _, dir := range subDirs { // FIXME: this should only remove things that have been removed
		if !dir.IsDir() {
			continue
		}
		remove := true
		for _, sub := range allowedSubDirs {
			if sub == dir.Name() {
				remove = false
				break
			}
		}
		if remove {
			err := os.RemoveAll(parent + "/" + dir.Name())
			if err != nil {
				log.Error(err)
			}
		}
	}
}
