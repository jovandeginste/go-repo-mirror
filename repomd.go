package main

import (
	"encoding/xml"
	"fmt"
	"path"
)

// RepoMetadata contains the outermost repository metadata
type RepoMetadata struct {
	XMLName     xml.Name
	Origin      string
	Destination string
	Revision    string      `xml:"revision"`
	Data        []*RepoData `xml:"data"`
}

// Checksum is a generic struct used in the metadata
type Checksum struct {
	Type  string `xml:"type,attr"`
	PkgID string `xml:"pkgid,attr"`
	Text  string `xml:",chardata"`
}

// RepoLocation refers to the resouce location relative to the repository root
type RepoLocation struct {
	Href string `xml:"href,attr"`
}

// RepoData contains metadata about a secondary level of repository metadata resources
type RepoData struct {
	Origin          string
	Destination     string
	Type            string       `xml:"type,attr"`
	Checksum        Checksum     `xml:"checksum"`
	Location        RepoLocation `xml:"location"`
	Timestamp       float64      `xml:"timestamp"`
	Size            int64        `xml:"size"`
	OpenChecksum    Checksum     `xml:"open-checksum,omitempty"`
	OpenSize        int64        `xml:"open-size,omitempty"`
	DatabaseVersion int          `xml:"database_version,omitempty"`
}

func (r *RepoMetadata) findRepoData(dataType string) *RepoData {
	for _, data := range r.Data {
		if data.Type == dataType {
			return data
		}
	}
	return nil
}

// URL returns the full URL for the resource
func (r *RepoData) URL() string {
	return urlJoin(r.Origin, r.Location.Href)
}

// Fetch will return the resource's content
func (r *RepoData) Fetch() ([]byte, error) {
	if r.Location.Href == "" {
		return []byte{}, fmt.Errorf("this element has no location href")
	}
	if r.Origin == "" {
		return []byte{}, fmt.Errorf("this element has no origin")
	}

	content := unpackURL(r.URL())
	return content, nil
}

// Primary will return the metadata resource with the primary package information
func (r *RepoMetadata) Primary() *RepoData {
	return r.findRepoData("primary")
}

// Filelists will return the metadata resource with the file list information
func (r *RepoMetadata) Filelists() *RepoData {
	return r.findRepoData("filelists")
}

// FetchPackages downloads all package metadata (using the Primary data resource)
func (r *RepoMetadata) FetchPackages() []*RepoPackage {
	data, err := r.Primary().Fetch()
	if err != nil {
		logItFatal(err)
	}
	meta := RepoPackageMeta{Origin: r.Origin, Destination: r.Destination}
	err = xml.Unmarshal(data, &meta)
	if err != nil {
		logItFatal(err)
	}
	for _, p := range meta.Package {
		p.Origin = r.Origin
		p.Destination = r.Destination
	}
	logItf(1, "We have %d packages.", len(meta.Package))

	return meta.Package
}

// FileLocation returns the location on disk
func (r *RepoData) FileLocation() string {
	return path.Join(r.Destination, r.Location.Href)
}

// MirrorMetadata will mirror all metadata resources
func (r *RepoMetadata) MirrorMetadata(sizeCheck bool, concurrent int) {
	data := r.Data

	for _, d := range data {
		d.MirrorCond(sizeCheck)
	}
}

// MirrorRepomd will download the repomd.xml file
func (r *RepoMetadata) MirrorRepomd(sizeCheck bool, concurrent int) {
	source := r.RepomdURL()
	destination := r.RepomdFileLocation()

	makeDestination(destination)

	writeURLToFile(source, destination)
}

// MirrorCond will conditionally mirror a metadata resource
func (r *RepoData) MirrorCond(sizeCheck bool) bool {
	ma := mirrorAction{
		Name:         r.Location.Href,
		Source:       r.URL(),
		Destination:  r.FileLocation(),
		ChecksumType: r.Checksum.Type,
		Checksum:     r.Checksum.Text,
		Size:         r.Size,
	}

	return ma.mirrorCond(sizeCheck)
}

// MirrorPackages will mirror all packages
func (r *RepoMetadata) MirrorPackages(sizeCheck bool, concurrent int) {
	packages := r.FetchPackages()
	repoChan := make(chan *RepoPackage, concurrent)
	resultChan := make(chan bool)

	for i := 0; i < concurrent; i++ {
		go func() {
			var p *RepoPackage
			var ok bool
			for {
				p, ok = <-repoChan
				if !ok {
					return
				}

				p.MirrorCond(sizeCheck)
				resultChan <- true
			}
		}()
	}

	go func() {
		for _, p := range packages {
			repoChan <- p
		}
		close(repoChan)
	}()

	results := 0
	totalPackages := len(packages)
	for {
		if results == totalPackages {
			break
		}
		<-resultChan
		results++
		if results%1000 == 0 {
			logItf(2, "Received packages: %d/%d", results, totalPackages)
		}
	}
	logItf(1, "Received: %d/%d", results, totalPackages)
	close(resultChan)
}

// RepomdURL returns the full URL to the repomd.xml file
func (r *RepoMetadata) RepomdURL() string {
	return urlJoin(r.Origin, "repodata/repomd.xml")
}

// RepomdFileLocation returns the full location on disk of the repomd.xml file
func (r *RepoMetadata) RepomdFileLocation() string {
	return path.Join(r.Destination, "repodata/repomd.xml")
}
