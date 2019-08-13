package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ApiAddr string
	OpsAddr string
}

var config = Config{
	ApiAddr: "127.0.0.1:8080",
	OpsAddr: "127.0.0.1:8081",
}

func ConfigFromPath(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
