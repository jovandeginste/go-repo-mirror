package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	verbose             = kingpin.Flag("verbose", "Verbosity level (0 to silence).").Short('v').Default("1").Int()
	logFile             = kingpin.Flag("log-file", "File to write logs to (logs still go to stdout).").Short('l').String()
	metadataOnly        = kingpin.Flag("metadata-only", "Only download repository metadata.").Short('m').Bool()
	dataOnly            = kingpin.Flag("data-only", "Only download repository data.").Short('d').Bool()
	concurrentDownloads = kingpin.Flag("concurrent-downloads", "Number of concurrent downloads.").Short('c').Default("10").Int()
	sizeCheck           = kingpin.Flag("size-check", "Don't verify file hash.").Bool()
	certFile            = kingpin.Flag("cert", "Client certificate file (PEM).").String()
	keyFile             = kingpin.Flag("key", "Client private key file (PEM).").String()
	insecureTLS         = kingpin.Flag("insecure-tls", "Disable TLS check for server.").Bool()
	dataPath            = kingpin.Flag("data-path", "Path to store the data(if not inside the destination folder).").String()
	metadataPath        = kingpin.Flag("metadata-path", "Path to store the metadata(if not inside the destination folder).").String()
	head                = kingpin.Flag("head", "Don't fetch the full package but only use a HEAD request.").Bool()

	repoURL           = kingpin.Arg("repo-url", "Remote URL to mirror the repository from.").Required().String()
	destinationFolder = kingpin.Arg("destination-folder", "Local folder to mirror the repository to.").Required().String()
)

func main() {
	kingpin.Parse()

	if *logFile != "" {
		makeDestination(*logFile)
		file, err := os.Create(*logFile)
		if err != nil {
			panic(err)
		}
		loggers = append(loggers, log.New(file, "", log.LstdFlags))
	}
	loggers = append(loggers, log.New(os.Stdout, "", log.LstdFlags))

	if !strings.HasSuffix(*repoURL, "/") {
		*repoURL = *repoURL + "/"
	}

	if *dataPath == "" {
		dataPath = destinationFolder
	}
	if *metadataPath == "" {
		metadataPath = destinationFolder
	}

	if !strings.HasSuffix(*dataPath, "/") {
		*dataPath = *dataPath + "/"
	}
	if !strings.HasSuffix(*metadataPath, "/") {
		*metadataPath = *metadataPath + "/"
	}
	logIt(0, "Mirroring '"+*repoURL+"' data to '"+*dataPath+"'.")
	logIt(0, "Mirroring '"+*repoURL+"' metadata to '"+*metadataPath+"'.")

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
			ResponseHeaderTimeout: 300 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tlsConfig,
		},
	}

	SetClient(&c)
	data := RepositoryMirror(*repoURL, *metadataPath, *dataPath)

	if !*metadataOnly {
		logIt(0, "Downloading packages...")
		data.MirrorPackages(*sizeCheck, *concurrentDownloads)
	}
	if !*dataOnly {
		logIt(0, "Downloading metadata...")
		data.MirrorMetadata(*sizeCheck, *concurrentDownloads)
		logIt(0, "Downloading repomd.xml...")
		data.MirrorRepomd(*sizeCheck, *concurrentDownloads)
	}
}
