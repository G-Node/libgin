package libgin

import (
	"encoding/xml"
	"strings"
	"time"
)

const (
	Schema         = "http://www.w3.org/2001/XMLSchema-instance"
	Namespace      = "http://datacite.org/schema/kernel-4"
	SchemaLocation = "http://datacite.org/schema/kernel-4 http://schema.datacite.org/meta/kernel-4/metadata.xsd"
	Publisher      = "G-Node"
	Language       = "eng"
	Version        = "1.0"
)

type NameIdentifier struct {
	ID        string `xml:",chardata"`
	SchemeURI string `xml:"schemeURI,attr"`
	Scheme    string `xml:"nameIdentifierScheme,attr"`
}

type Creator struct {
	Name        string          `xml:"creatorName"`
	Identifier  *NameIdentifier `xml:"nameIdentifier,omitempty"`
	Affiliation string          `xml:"affiliation,omitempty"`
}

type Description struct {
	Content string `xml:",chardata"`
	Type    string `xml:"descriptionType,attr"`
}

type Rights struct {
	Name string `xml:",chardata"`
	URL  string `xml:"rightsURI,attr"`
}

type RelatedIdentifier struct {
	Identifier   string `xml:",chardata"`
	Type         string `xml:"relatedIdentifierType,attr"`
	RelationType string `xml:"relationType,attr"`
}

type FundingReference struct {
	Funder      string `xml:"funderName"`
	AwardNumber string `xml:"awardNumber"`
	// TODO: Add identifier for known funders
}

type Contributor struct {
	Name string `xml:"contributorName"`
	Type string `xml:"contributorType,attr"`
}

type Date struct {
	Value string `xml:",chardata"`
	// Always set to "Issued"
	Type string `xml:"dateType,attr"`
}

type ResourceType struct {
	Value   string `xml:",chardata"`
	General string `xml:"resourceTypeGeneral,attr"`
}
type DataCite struct {
	XMLName        xml.Name `xml:"resource"`
	Schema         string   `xml:"xmlns:xsi,attr"`
	Namespace      string   `xml:"xmlns,attr"`
	SchemaLocation string   `xml:"xsi:schemaLocation,attr"`
	// Creators: Authors
	Creators     []Creator     `xml:"creators>creator"`
	Titles       []string      `xml:"titles>title"`
	Descriptions []Description `xml:"descriptions>description"`
	// RightsList: Licenses
	RightsList Rights `xml:"rightsList>rights"`
	// Subjects: Keywords
	Subjects []string `xml:"subjects>subject,omitempty"`
	// RelatedIdentifiers: References
	RelatedIdentifiers []RelatedIdentifier `xml:"relatedIdentifiers,omitempty>relatedIdentifier"`
	FundingReferences  []FundingReference  `xml:"fundingReferences>fundingReference,omitempty"`
	// Contributors: Always German Neuroinformatics Node with type "HostingInstitution"
	Contributors []Contributor `xml:"contributors>contributor"`
	// Publisher: Always G-Node
	Publisher string `xml:"publisher"`
	// Publication Year
	Year int `xml:"publicationYear"`
	// Publication Date marked with type "Issued"
	Dates []Date `xml:"dates>date"`
	// Language: eng
	Language     string       `xml:"language"`
	ResourceType ResourceType `xml:"resourceType"`
	// Version: 1.0
	Version string `xml:"version"`
}

// Returns a DataCite struct populated with our defaults.
// The following values are set and generally shouldn't be changed:
// Schema, Namespace, SchemaLocation, Contributors, Publisher, Language, Version.
// Dates and Year are also pre-filled with the current date but should be
// changed when working with an existing publication.
func NewDataCite() DataCite {
	return DataCite{
		Schema:         Schema,
		Namespace:      Namespace,
		SchemaLocation: SchemaLocation,
		Contributors:   []Contributor{Contributor{"German Neuroinformatics Node", "HostingInstitution"}},
		Publisher:      Publisher,
		Year:           time.Now().Year(),
		Dates:          []Date{Date{time.Now().Format("2006-01-02"), "Issued"}},
		Language:       Language,
		Version:        Version,
	}
}

// SetResourceType is a convenience function for setting the ResourceType data
// and its resourceTypeGeneral to the same value.
func (dc *DataCite) SetResourceType(resourceType string) {
	dc.ResourceType = ResourceType{resourceType, resourceType}
}

// AddFunding is a convenience function for appending a FundingReference in the
// format of the YAML data (<FUNDER>, <AWARDNUMBER>).
func (dc *DataCite) AddFunding(fundstr string) {
	funParts := strings.SplitN(fundstr, ",", 2)
	var funder, awardNumber string
	if len(funParts) == 2 {
		funder = strings.TrimSpace(funParts[0])
		awardNumber = strings.TrimSpace(funParts[1])
	} else {
		// No comma, add to award number as is
		awardNumber = fundstr
	}
	fundref := FundingReference{Funder: funder, AwardNumber: awardNumber}
	dc.FundingReferences = append(dc.FundingReferences, fundref)
}
