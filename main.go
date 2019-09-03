package main

import (
	"reflect"

	"github.com/microsoft/fabrikate/cmd"
	"github.com/timfpark/yaml"
)

func main() {
	// modify the DefaultMapType of yaml to map[string]interface{} instead of map[interface]interface{}
	*yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})

	cmd.Execute()
}
