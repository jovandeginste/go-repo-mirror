package main

import (
	"encoding/xml"
	"log"
	"os"
)

type mirrorAction struct {
	Name         string
	Source       string
	Destination  string
	ChecksumType string
	Checksum     string
	Size         int64
}

// RepositoryMirror initializes a mirror between a remote url and a local folder
func RepositoryMirror(repoURL string, destination string, dataDestination string) *RepoMetadata {
	repomd := RepoMetadata{Origin: repoURL, Destination: destination, DataDestination: dataDestination}
	url := repomd.RepomdURL()
	logIt(1, "Fetching repomd url '"+url+"'")
	result := getURL(url)

	err := xml.Unmarshal(result, &repomd)
	if err != nil {
		logItFatal(err)
	}
	for _, data := range repomd.Data {
		data.Origin = repoURL
		data.Destination = destination
	}
	return &repomd
}

func (r *mirrorAction) mirrorCond(sizeCheck bool) bool {
	if r.verify(sizeCheck) {
		return true
	}
	logIt(1, "We don't have the correct version of '"+r.Name+"'. Downloading it...")
	r.mirror()
	return false
}

func (r *mirrorAction) mirror() string {
	source := r.Source
	destination := r.Destination

	tmpDestination := destination + ".part"

	makeDestination(tmpDestination)

	writeURLToFile(source, tmpDestination)
	if !r.verifyChecksumPath(tmpDestination) {
		log.Fatal("WTF?")
	}
	os.Rename(tmpDestination, destination)
	return destination
}

func (r *mirrorAction) verify(sizeCheck bool) bool {
	if sizeCheck {
		return r.verifySize()
	}
	return r.verifyChecksum()
}

func (r *mirrorAction) verifySize() bool {
	return r.verifySizePath(r.Destination)
}

func (r *mirrorAction) verifySizePath(path string) bool {
	remote := r.Size
	local := sizeForFile(path)
	return remote == local
}

func (r *mirrorAction) verifyChecksum() bool {
	return r.verifyChecksumPath(r.Destination)
}

func (r *mirrorAction) verifyChecksumPath(path string) bool {
	remote := r.Checksum
	local := checksumForFile(path, r.ChecksumType)
	return remote == local
}
