package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/leafo/zip_server"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type gsURL struct {
	Bucket string
	Key    string
}

var (
	configFname         string
	maxBytes            int
	acl                 string
	contentDisposition  string
	contentTypeOverride string
)

func init() {
	flag.StringVar(&configFname, "config", defaultConfigFname, "Path to json config file")
	flag.IntVar(&maxBytes, "max_bytes", 0, "Max bytes to copy (0 is no limit)")
	flag.StringVar(&acl, "acl", "public-read", "ACL of uploaded file")
	flag.StringVar(&contentDisposition, "content_disposition", "", "Content disposition of uploaded file")
	flag.StringVar(&contentTypeOverride, "content_type", "", "Content type of uploaded file (defaults to content type from HTTP request)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: url2gs [OPTIONS] http://URL gs://BUCKET/KEY\n\nOptions:\n")
		flag.PrintDefaults()
	}
}

type limitedReader func(p []byte) (int, error)

func (fn limitedReader) Read(p []byte) (int, error) {
	return fn(p)
}

// wraps reader to fail if it reads too many bytes
func newLimitedReader(reader io.Reader, maxBytes int) limitedReader {
	remainingBytes := maxBytes
	return func(p []byte) (int, error) {
		bytesRead, err := reader.Read(p)
		remainingBytes -= bytesRead

		if remainingBytes < 0 {
			return bytesRead, fmt.Errorf("limited reader passed limit %d", maxBytes)
		}

		return bytesRead, err
	}
}

func parseGsURL(url string) (gsURL, error) {
	patt := regexp.MustCompile("^gs://([^/]+)/(.*)$")
	match := patt.FindStringSubmatch(url)

	if len(match) == 0 {
		return gsURL{}, errors.New("invalid gs:// URL syntax: " + url)
	}

	return gsURL{
		Bucket: match[1],
		Key:    match[2],
	}, nil
}

func main() {
	flag.Parse()
	config := loadConfig(configFname)

	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("missing URL")
	}

	if len(args) < 2 {
		log.Fatal("missing Cloud Storage URL")
	}

	target, err := parseGsURL(args[1])

	if err != nil {
		log.Fatal(err.Error())
	}

	storage := zip_server.StorageClient{
		PrivateKeyPath: config.PrivateKeyPath,
		ClientEmail:    config.ClientEmail,
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

	contentType := contentTypeOverride

	if contentType == "" {
		contentType = res.Header.Get("Content-Type")
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var body io.Reader = res.Body

	contentLengthStr := res.Header.Get("Content-Length")
	if maxBytes > 0 {
		log.Print("setting max size to ", maxBytes)
		if contentLengthStr != "" {
			contentLength, err := strconv.Atoi(contentLengthStr)

			if err != nil {
				log.Fatal("invalid content length from response")
			}

			if contentLength > maxBytes {
				log.Fatal("content length greater than max size (", contentLength, " > ", maxBytes, ")")
			}
		}

		body = newLimitedReader(body, maxBytes)
	}

	log.Print("Uploading ", contentType, " (size: ", contentLengthStr, ") to ", target.Key)
	log.Print("ACL: ", acl)
	log.Print("Content-Disposition: ", contentDisposition)

	err = storage.PutFileWithSetup(target.Bucket, target.Key, body, func(req *http.Request) error {
		req.Header.Add("Content-Type", contentType)
		if contentDisposition != "" {
			req.Header.Add("Content-Disposition", contentDisposition)
		}

		req.Header.Add("x-goog-acl", acl)
		return nil
	})

	if err != nil {
		log.Fatal(err.Error())
	}
}
