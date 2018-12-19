package main

import (
	"github.com/Microsoft/fabrikate/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true

	log.SetFormatter(formatter)
	log.SetLevel(log.InfoLevel)

	cmd.Execute()
}
