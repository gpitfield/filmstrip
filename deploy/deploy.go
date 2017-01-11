package deploy

import (
	"io/ioutil"
	"time"

	"github.com/gpitfield/filmstrip/deploy/driver"
	_ "github.com/gpitfield/filmstrip/deploy/driver/drivers/s3"
	"github.com/gpitfield/filmstrip/site"
	log "github.com/gpitfield/relog"
	"github.com/spf13/viper"
)

const (
	SITE_DRIVER = "driver"
)

type PutJob struct {
	localPrefix string
	path        string
	force       bool
}

// Deploy the site to its destination, forcing overwrite if force is true
func Deploy(force bool) {
	start := time.Now()
	DeployDirs(site.PubSiteDir, "", force, nil, nil)
	log.Infof("site deployed in %v", time.Since(start))
	start = time.Now()
	Flush() // remove any files that no longer belong
	log.Infof("site flushed in %v", time.Since(start))
}

func Flush() {
	drv := viper.GetString(SITE_DRIVER)
	paths := GetPaths("")
	err := driver.Drivers[drv].FlushFiles(paths)
	if err != nil {
		log.Error(err)
	}
}

func GetPaths(prefix string) (paths []string) {
	files, err := ioutil.ReadDir(site.PubSiteDir + prefix)
	if err != nil {
		log.Error(err)
	}
	for _, f := range files {
		if f.IsDir() {
			paths = append(paths, GetPaths(prefix+"/"+f.Name())...)
			continue
		}
		paths = append(paths, prefix+"/"+f.Name())
	}
	return paths
}

func DeployDirs(localPrefix string, path string, force bool, jobs chan PutJob, done chan bool) {
	closer := false
	if jobs == nil {
		closer = true
		done = make(chan bool)
		workers := viper.GetInt("workers")
		if workers == 0 {
			workers = 1
		}
		jobs = make(chan PutJob)
		log.Infof("Deploying on %d workers", workers)
		for i := 0; i < workers; i++ {
			go PutFiles(jobs, done)
		}
	}

	files, err := ioutil.ReadDir(localPrefix + path)
	if err != nil {
		log.Error(err)
	}
	for _, f := range files {
		if f.IsDir() {
			DeployDirs(localPrefix, path+"/"+f.Name(), force, jobs, done)
			continue
		}
		jobs <- PutJob{localPrefix, path + "/" + f.Name(), force}
	}
	if closer {
		close(jobs)
		workers := viper.GetInt("workers")
		for i := 0; i < workers; i++ {
			<-done
		}
	}
}

func PutFiles(jobs chan PutJob, done chan bool) {
	var (
		err error
		drv = viper.GetString(SITE_DRIVER)
	)
	if drv == "" {
		log.Warnf("please set a value for %s in the config file", SITE_DRIVER)
		return
	}

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				done <- true
				return
			}
			err = driver.Drivers[drv].PutFile(job.localPrefix, job.path, job.force)
			if err != nil {
				log.Errorf("%s %s", job.localPrefix+job.path, err.Error())
			}
		}
	}
}
