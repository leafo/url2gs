package main

import (
	"fmt"
	"log"
	"os"
	"flag"
	"regexp"
	"errors"
	. "github.com/leafo/url2gs"
	"github.com/leafo/zip_server"
)


type GsUrl struct {
	Bucket string
	Key string
}

var (
	configFname string
	maxBytes int
)

var usage = "\nUsage: url2gs [-max_bytes=MAX] http://URL gs://BUCKET/KEY"

func init() {
	flag.StringVar(&configFname, "config", DefaultConfigFname, "Path to json config file")
	flag.IntVar(&maxBytes, "max_bytes", 0, "Max bytes to copy")
}

func ParseGsUrl(url string) (GsUrl, error) {
	patt := regexp.MustCompile("^gs://([^/]+)/(.*)$")
	match := patt.FindStringSubmatch(url)
	fmt.Println("got matches", len(match))

	if len(match) == 0 {
		return GsUrl{}, errors.New("invalid gs:// URL syntax: " + url)
	}

	return GsUrl{
		Bucket: match[1],
		Key: match[2],
	}, nil
}

func main() {
	flag.Parse()
	config := LoadConfig(configFname)

	if len(os.Args) < 2 {
		log.Fatal("missing URL" + usage)
	}

	if len(os.Args) < 3 {
		log.Fatal("missing Cloud Storage URL" + usage)
	}

	target, err := ParseGsUrl(os.Args[2])

	if err != nil {
		log.Fatal(err.Error() + usage)
	}

	storage := zip_server.StorageClient{
		PrivateKeyPath: config.PrivateKeyPath,
		ClientEmail: config.ClientEmail,
	}

	fmt.Println("okay:", target, storage)
}

