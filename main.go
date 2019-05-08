package main

import (
	"github.com/microsoft/fabrikate/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true

	log.SetFormatter(formatter)

	cmd.Execute()
}
