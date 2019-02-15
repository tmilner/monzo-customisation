package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Configuration struct {
	Authorization string `json:"auth"`
}

func New() *Configuration {
	contents, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println("Config error!", err)
	}

	var configuration Configuration
	err = json.Unmarshal(contents, &configuration)
	if err != nil {
		log.Println("A different config error:", err)
	}
	return &configuration
}
