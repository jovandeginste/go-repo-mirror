package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	verbose             = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	repoURL             = kingpin.Arg("repo-url", "Remote URL to mirror the repository from.").Required().String()
	destinationFolder   = kingpin.Arg("destination-folder", "Local folder to mirror the repository to.").Required().String()
	metadataOnly        = kingpin.Flag("metadata-only", "Only download repository metadata.").Short('m').Bool()
	dataOnly            = kingpin.Flag("data-only", "Only download repository data.").Short('d').Bool()
	concurrentDownloads = kingpin.Flag("concurrent-downloads", "Number of concurrent downloads.").Short('c').Default("10").Int()
	sizeCheck           = kingpin.Flag("size-check", "Don't verify file hash.").Bool()
	certFile            = kingpin.Flag("cert", "Client certificate file (PEM).").String()
	keyFile             = kingpin.Flag("key", "Client private key file (PEM).").String()
	insecureTLS         = kingpin.Flag("insecure-tls", "Disable TLS check for server.").Bool()
)

func main() {
	kingpin.Parse()
	if !strings.HasSuffix(*repoURL, "/") {
		*repoURL = *repoURL + "/"
	}
	if !strings.HasSuffix(*destinationFolder, "/") {
		*destinationFolder = *destinationFolder + "/"
	}
	logIt("Mirroring '" + *repoURL + "' to '" + *destinationFolder + "'.")

	tlsConfig := tls.Config{
		InsecureSkipVerify: *insecureTLS,
	}

	if *certFile != "" && *keyFile != "" {
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			logItFatal(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	c := http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tlsConfig,
		},
	}

	SetClient(&c)
	data := RepositoryMirror(*repoURL, *destinationFolder)

	if !*metadataOnly {
		logIt("Downloading packages...")
		data.MirrorPackages(*sizeCheck, *concurrentDownloads)
	}
	if !*dataOnly {
		logIt("Downloading metadata...")
		data.MirrorMetadata(*sizeCheck, *concurrentDownloads)
		logIt("Downloading repomd.xml...")
		data.MirrorRepomd(*sizeCheck, *concurrentDownloads)
	}
}
