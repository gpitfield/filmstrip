package main

import (
	"os"

	"github.com/gpitfield/filmstrip/cmd"
	log "github.com/gpitfield/relog"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}
