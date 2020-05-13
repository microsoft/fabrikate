package yaml

import (
	"github.com/microsoft/fabrikate/pkg/logger"
	"github.com/timfpark/yaml"
	"reflect"
)

func Marshal(in interface{}) (out []byte, err error) {
	return yaml.Marshal(in)
}

func Unmarshal(in []byte, out interface{}) (err error) {
	return yaml.Unmarshal(in, out)
}

func init() {
	logger.Debug("initializing yaml")
	// modify the DefaultMapType of yaml to map[string]interface{} instead of map[interface]interface{}
	*yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
	logger.Debug("setting defaultMapType to map[string]interface{}")
}
