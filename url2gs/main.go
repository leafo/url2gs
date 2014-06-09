package main

import (
	"fmt"
	"strconv"
	"log"
	"io"
	"flag"
	"regexp"
	"errors"
	. "github.com/leafo/url2gs"
	"github.com/leafo/zip_server"
	"net/http"
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

type LimitedReader func (p []byte) (int, error)

func (fn LimitedReader) Read(p []byte) (int, error) {
	return fn(p)
}

// wraps reader to fail if it reads too many bytes
func NewLimitedReader(reader io.Reader, maxBytes int) LimitedReader {
	remainingBytes := maxBytes
	return func (p []byte) (int, error) {
		bytesRead, err := reader.Read(p)
		remainingBytes -= bytesRead

		if remainingBytes < 0 {
			return bytesRead, fmt.Errorf("limited reader passed limit %d", maxBytes)
		}

		return bytesRead, err
	}
}

func ParseGsUrl(url string) (GsUrl, error) {
	patt := regexp.MustCompile("^gs://([^/]+)/(.*)$")
	match := patt.FindStringSubmatch(url)

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

	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("missing URL" + usage)
	}

	if len(args) < 2 {
		log.Fatal("missing Cloud Storage URL" + usage)
	}

	target, err := ParseGsUrl(args[1])

	if err != nil {
		log.Fatal(err.Error() + usage)
	}

	storage := zip_server.StorageClient{
		PrivateKeyPath: config.PrivateKeyPath,
		ClientEmail: config.ClientEmail,
	}

	client := http.Client{}
	res, err := client.Get(args[0])
	defer res.Body.Close()

	if err != nil {
		log.Fatal("failed to create http client: " + err.Error())
	}

	if res.StatusCode != 200 {
		log.Fatal("failed to fetch file, status: ", res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	contentLength, err := strconv.Atoi(res.Header.Get("Content-Length"))

	if err != nil {
		log.Fatal("missing content length from response")
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var body io.Reader = res.Body

	if maxBytes > 0 {
		log.Print("setting max size to ", maxBytes)

		if contentLength > maxBytes {
			log.Fatal("content length greater than max size (", contentLength, " > ", maxBytes, ")")
		}

		body = NewLimitedReader(body, maxBytes)
	}

	log.Print("Uploading ", contentType, " (size: ", res.Header.Get("Content-Length"), ") to ", target.Key)

	err = storage.PutFile(target.Bucket, target.Key, body, contentType)

	if err != nil {
		log.Fatal(err.Error())
	}
}

