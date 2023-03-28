package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func ParseToServicesConfig(filepath string) (*ServicesConfig, error) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Printf("common ParseToStruct: filepath %s wrong", filepath)
		return nil, err
	}

	conf := &ServicesConfig{}
	err = yaml.Unmarshal(file, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
