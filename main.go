package main

import (
	"reflect"

	"github.com/microsoft/fabrikate/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/timfpark/yaml"
)

func main() {
	// modify the DefaultMapType of yaml to map[string]interface{} instead of map[interface]interface{}
	*yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})

	formatter := new(log.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true

	log.SetFormatter(formatter)

	cmd.Execute()
}
