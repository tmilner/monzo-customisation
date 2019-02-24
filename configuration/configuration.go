package configuration

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	Authorization string `json:"auth"`
}

func New() (*Configuration, error) {
	contents, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var configuration Configuration
	err = json.Unmarshal(contents, &configuration)

	return &configuration, err
}
