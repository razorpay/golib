package config

import (
	"encoding/json"
)

func Parse(configInString map[string]interface{}, cfgStructPointer interface{}) error {
	if configInString == nil {
		return nil
	}
	b, err := json.Marshal(configInString)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, cfgStructPointer)
	if err != nil {
		return err
	}
	return nil
}
