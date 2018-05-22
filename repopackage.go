package main

import (
	"encoding/xml"
	"path"
)

// RepoPackageMeta is the structure containing RepoPackage metadata
type RepoPackageMeta struct {
	XMLName     xml.Name
	Origin      string
	Destination string
	Package     []*RepoPackage `xml:"package"`
}

// RepoPackage contains a subset of the package metadata
type RepoPackage struct {
	Origin      string
	Destination string
	Name        string          `xml:"name"`
	Arch        string          `xml:"arch"`
	Version     PackageVersion  `xml:"version"`
	Checksum    Checksum        `xml:"checksum"`
	Summary     string          `xml:"summary"`
	Description string          `xml:"description"`
	Package     string          `xml:"packager"`
	ProjectURL  string          `xml:"url"`
	Time        RepoPackageTime `xml:"time"`
	Size        RepoPackageSize `xml:"size"`
	Location    RepoLocation    `xml:"location"`
}

// RepoPackageSize is a number of size attributes
type RepoPackageSize struct {
	Package   int64 `xml:"package,attr"`
	Installed int64 `xml:"installed,attr"`
	Archive   int64 `xml:"archive,attr"`
}

// RepoPackageTime is a number of time attributes
type RepoPackageTime struct {
	File  int `xml:"file,attr"`
	Build int `xml:"build,attr"`
}

// PackageVersion is the full (structured) version of the package
type PackageVersion struct {
	Epoch string `xml:"epoch,attr"`
	Ver   string `xml:"ver,attr"`
	Rel   string `xml:"rel,attr"`
}

// URL returns the full URL for the resource
func (p *RepoPackage) URL() string {
	return urlJoin(p.Origin, p.Location.Href)
}

// FileLocation returns the location on disk
func (p *RepoPackage) FileLocation() string {
	return path.Join(p.Destination, p.Location.Href)
}

// MirrorCond will conditionally mirror a metadata resource
func (p *RepoPackage) MirrorCond(sizeCheck bool) bool {
	ma := mirrorAction{
		Name:         p.Location.Href,
		Source:       p.URL(),
		Destination:  p.FileLocation(),
		ChecksumType: p.Checksum.Type,
		Checksum:     p.Checksum.Text,
		Size:         p.Size.Package,
	}

	return ma.mirrorCond(sizeCheck)
}
