package main

import (
	"encoding/xml"
	"io"
)

// Specification at: https://github.com/robertmaldon/cc_dashboard#summary

// Project represents a phase in a build pipeline that is reported on
type Project struct {
	Name            string          `xml:"name,attr"`
	Activity        Activity        `xml:"activity,attr"`
	LastBuildLabel  string          `xml:"lastBuildLabel,attr,omitempty"`
	LastBuildStatus LastBuildStatus `xml:"lastBuildStatus,attr"`
	LastBuildTime   string          `xml:"lastBuildTime,attr"`
	NextBuildTime   string          `xml:"nextBuildTime,attr,omitempty"`
	WebURL          string          `xml:"webUrl,attr"`
}

type projectsContainer struct {
	XMLName  xml.Name  `xml:"Projects"`
	Projects []Project `xml:"Project"`
}

// Encode the projects as XML
func Encode(projects []Project, w io.Writer) error {
	return xml.NewEncoder(w).Encode(projectsContainer{Projects: projects})
}

// LastBuildStatus describes the status of the most recent build
type LastBuildStatus string

const (
	// LastBuildStatusSuccess is the string used to indicate that the most recent build succeeded
	LastBuildStatusSuccess LastBuildStatus = "Success"
	// LastBuildStatusFailure is the string used to indicate that the most recent build failed
	LastBuildStatusFailure LastBuildStatus = "Failure"
	// LastBuildStatusException is the string used to indicate that the most recent build had an exception
	LastBuildStatusException LastBuildStatus = "Exception"
	// LastBuildStatusUnknown is the string used to indicate that the most recent build status is unknown
	LastBuildStatusUnknown LastBuildStatus = "Unknown"
)

// Activity describes the current status of a build
type Activity string

const (
	// ActivityBuilding is the string used to indicate that the project is currently building
	ActivityBuilding Activity = "Building"
	// ActivitySleeping is the string used to indicate that the project is currently sleeping
	ActivitySleeping Activity = "Sleeping"
	// ActivityCheckingModifications is the string used to indicate that the project is checking for modifications
	ActivityCheckingModifications Activity = "CheckingModifications"
)
