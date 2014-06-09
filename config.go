package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var defaultConfigFname = "url2gs.json"

type config struct {
	PrivateKeyPath string
	ClientEmail    string
	Bucket         string
}

func loadConfig(fname string) *config {
	jsonBlob, err := ioutil.ReadFile(fname)

	if err != nil {
		log.Fatal(err)
	}

	c := config{}
	err = json.Unmarshal(jsonBlob, &c)

	if err != nil {
		log.Fatal("Failed parsing config: " + fname + ": " + err.Error())
	}

	return &c
}
