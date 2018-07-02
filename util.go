package main

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/ulikunitz/xz"
)

var client *http.Client
var clientInit = false

// SetClient let's you configure a custop http Client
func SetClient(c *http.Client) {
	client = c
	clientInit = true
}

func initClient() {
	c := http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	SetClient(&c)
}

func urlJoin(base, refpath string) string {
	url1, err := url.Parse(base)
	if err != nil {
		logItFatal(err)
	}
	url2, err := url.Parse(refpath)
	if err != nil {
		logItFatal(err)
	}
	return url1.ResolveReference(url2).String()
}

func getURL(url string) []byte {
	response, err := client.Get(url)
	if err != nil {
		logItFatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		logItFatal("Got an unexpected response code: " + response.Status)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logItFatal(err)
	}
	return body
}

func unpackURL(url string) []byte {
	content := getURL(url)
	var err error
	extension := path.Ext(url)
	switch extension {
	case ".gz":
		content, err = gUnzipData(content)
		if err != nil {
			logItFatal(err)
		}
	case ".bz2":
		content, err = bUnzipData(content)
		if err != nil {
			logItFatal(err)
		}
	case ".xz":
		content, err = xUnzipData(content)
		if err != nil {
			logItFatal(err)
		}
	}
	return content
}

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func bUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r = bzip2.NewReader(b)

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func writeURLToFile(source string, destination string) {
	if !clientInit {
		initClient()
	}
	resp, err := client.Get(source)
	if err != nil {
		logItFatal(err)
	}
	defer resp.Body.Close()

	file, err := os.Create(destination)
	if err != nil {
		logItFatal(err)
	}
	defer file.Close()

	n, err := io.Copy(file, resp.Body)

	logItf(2, "We wrote %d bytes from '%s' to '%s'", n, source, destination)
}

func sizeForFile(filepath string) int64 {
	fi, e := os.Stat(filepath)
	if e != nil {
		return -1
	}
	// get the size
	return fi.Size()
}

func checksumForFile(filePath string, csType string) string {
	var returnSHA1String string

	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return "does not exist"
	}
	if err != nil {
		logItFatal(err)
	}

	defer file.Close()

	hash := hashByType(csType)

	if _, err := io.Copy(hash, file); err != nil {
		logItFatal(err)
	}

	hashInBytes := hash.Sum(nil)

	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String
}

func hashByType(hashType string) hash.Hash {
	switch hashType {
	case "sha":
		return sha1.New()
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	case "md5":
		return md5.New()
	default:
		logItFatal("Unknown checksum: " + hashType)
		return nil
	}
}

func destinationExists(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

func makeDestination(path string) {
	dir := filepath.Dir(path)
	if !destinationExists(dir) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			logItFatal(err)
		}
	}
}

func xUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = xz.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}
