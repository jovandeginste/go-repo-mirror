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
	if !strings.HasSuffix(*destinationFolder, "/") {
		*destinationFolder = *destinationFolder + "/"
	}
	logIt(0, "Mirroring '"+*repoURL+"' to '"+*destinationFolder+"'.")

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
