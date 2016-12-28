package deploy

import (
	"time"

	log "github.com/gpitfield/relog"
)

// Deploy the site to its destination
func Deploy() {
	start := time.Now()
	log.Infof("Deployed in %v", time.Since(start))
}
