package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var DefaultConfigFname = "url2gs.json"

type Config struct {
	PrivateKeyPath string
	ClientEmail    string
	Bucket         string
}

func LoadConfig(fname string) *Config {
	jsonBlob, err := ioutil.ReadFile(fname)

	if err != nil {
		log.Fatal(err)
	}

	config := Config{}
	err = json.Unmarshal(jsonBlob, &config)

	if err != nil {
		log.Fatal("Failed parsing config: " + fname + ": " + err.Error())
	}

	return &config
}
