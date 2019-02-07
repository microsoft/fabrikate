package core

import (
	"fmt"
	"reflect"

	"github.com/InVisionApp/conjungo"
)

type ComponentConfig struct {
	Config        map[string]interface{}
	Subcomponents map[string]ComponentConfig
}

func coerceSliceValuesToStrings(configSlice []interface{}) (coercedSlice []interface{}, err error) {
	for idx, element := range configSlice {
		if element == nil {
			// nop: leave value as-is (nil)
		} else if reflect.TypeOf(element).Kind() == reflect.Bool {
			// nop: leave value as-is
		} else if reflect.TypeOf(element).Kind() == reflect.Map {
			configSlice[idx], err = coerceMapValuesToStrings(element.(map[string]interface{}))

			if err != nil {
				return configSlice, err
			}
		} else if reflect.TypeOf(element).Kind() == reflect.Slice {
			configSlice[idx], err = coerceSliceValuesToStrings(element.([]interface{}))

			if err != nil {
				return configSlice, err
			}
		} else {
			configSlice[idx] = fmt.Sprintf("%v", configSlice[idx])
		}
	}

	return configSlice, nil
}

func coerceMapValuesToStrings(configMap map[string]interface{}) (coercedMap map[string]interface{}, err error) {
	for key, value := range configMap {
		if value == nil {
			// nop: leave value as-is (nil)
		} else if reflect.TypeOf(value).Kind() == reflect.Bool {
			// nop: leave value as-is
		} else if reflect.TypeOf(value).Kind() == reflect.Map {
			configMap[key], err = coerceMapValuesToStrings(configMap[key].(map[string]interface{}))

			if err != nil {
				return configMap, err
			}
		} else if reflect.TypeOf(value).Kind() == reflect.Slice {
			configMap[key], err = coerceSliceValuesToStrings(configMap[key].([]interface{}))

			if err != nil {
				return configMap, err
			}
		} else {
			configMap[key] = fmt.Sprintf("%v", configMap[key])
		}
	}

	return configMap, nil
}

func (cc *ComponentConfig) CoerceConfigToStrings() (err error) {
	cc.Config, err = coerceMapValuesToStrings(cc.Config)

	if err != nil {
		return err
	}

	for _, subcomponent := range cc.Subcomponents {
		if err = subcomponent.CoerceConfigToStrings(); err != nil {
			return err
		}
	}

	return nil
}

func (cc *ComponentConfig) Merge(newConfig ComponentConfig) (err error) {
	options := conjungo.NewOptions()
	options.Overwrite = false

	return conjungo.Merge(cc, newConfig, options)
}
