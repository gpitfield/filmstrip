package deploy

import (
	"io/ioutil"

	"github.com/gpitfield/filmstrip/build"
	"github.com/gpitfield/filmstrip/deploy/driver"
	_ "github.com/gpitfield/filmstrip/deploy/driver/drivers/s3"
	log "github.com/gpitfield/relog"
	"github.com/spf13/viper"
)

const (
	SITE_DRIVER = "driver"
)

// Deploy the site to its destination, forcing overwrite if force is true
func Deploy(force bool) {
	DeployDir(2, "", force) // deploy up to 2 levels down
}

func DeployDir(levels int, prefix string, force bool) {
	drv := viper.GetString(SITE_DRIVER)
	files, err := ioutil.ReadDir(build.PubSiteDir + prefix)
	if err != nil {
		log.Error(err)
	}
	for _, f := range files {
		if f.IsDir() {
			if levels > 0 {
				DeployDir(levels-1, prefix+"/"+f.Name(), force)
			} else {
				log.Warnf("Sub-collections not supported; skipping %s", f.Name())
			}
			continue
		}
		err = driver.Drivers[drv].PutFile(build.PubSiteDir, prefix+"/"+f.Name(), force)
		if err != nil {
			log.Error(err)
		}
	}
}
