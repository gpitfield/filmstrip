package build

import (
	"strings"

	log "github.com/gpitfield/relog"
)

func checkErr(err error) {
	if err != nil && !strings.Contains(err.Error(), "file exists") {
		log.Error(err)
	}
}
