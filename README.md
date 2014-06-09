# url2gs

Uploads a file to Google Cloud storage from a HTTP URL.

If you just need a quick way to do it you can use the command line:

```bash
curl -f http://leafo.net/hi.png | gsutil -h "Content-Type:image/png" cp -a public-read - gs://leafo/hi.png
```

But if you want to handle HTTP status codes and limit the size of file that can
be uploaded then this tool is for you.


## Usage

Install

```bash
go get github.com/leafo/url2gs
go install github.com/leafo/url2gs/url2gs
```

Create a config file:

`url2gs.json`:

```json
{
	"PrivateKeyPath": "path/to/service/key.pem",
	"ClientEmail": "111111111111@developer.gserviceaccount.com"
}
```

Run:

```bash
$GOPATH/bin/url2gs -max_bytes=5000000 http://leafo.net/hi.png gs://leafo/hi.png
```

If the command succeeds then an exit code of 0 is returend, on failure 1 is
returned. Only requests that return a 200 status code are copied. The mime type
of the file is extracted from the `Content-Type` of the URL.


