package libgin

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	Schema         = "http://www.w3.org/2001/XMLSchema-instance"
	SchemaLocation = "http://datacite.org/schema/kernel-4 http://schema.datacite.org/meta/kernel-4.3/metadata.xsd"
	Publisher      = "G-Node"
	Language       = "eng"
	Version        = "1.0"
)

// relIDTypeMap is used to fix the case of reference types coming from the
// user's datacite.yml file.
var relIDTypeMap = map[string]string{
	"doi":   "DOI",
	"url":   "URL",
	"arxiv": "arXiv",
	"pmid":  "PMID",
}

// funderIDMap maps known funder names to their ID (DOI).
// This is currently populated from the existing registered datasets.
// A more comprehensive list may be added later.
var funderIDMap = map[string]string{
	"BMBF":                       "https://doi.org/10.13039/501100002347",
	"Boehringer Ingelheim Fonds": "https://doi.org/10.13039/501100001645",
	"CNRS":                       "https://doi.org/10.13039/501100004794",
	"DAAD":                       "https://doi.org/10.13039/501100001655",
	"DFG":                        "https://doi.org/10.13039/501100001659",
	"Einstein Foundation Berlin": "https://doi.org/10.13039/501100006188",
	"Einstein Stiftung":          "https://doi.org/10.13039/501100006188",
	"EU":                         "https://doi.org/10.13039/100010664",
	"European Union’s Horizon 2020 Framework Programme for Research and Innovation under the Specific Grant Agreements No. 720270 and No. 785907 (Human Brain Project SGA1 and SGA2; to M.M. and P.A.)": "https://doi.org/10.13039/100010661",
	"European Union's Seventh Framework Programme (FP/2007-2013)": "https://doi.org/10.13039/100011102",
	"Helmholtz Association":           "https://doi.org/10.13039/501100001656",
	"Human Frontiers Science Program": "https://doi.org/10.13039/501100000854",
	"Innovate UK":                     "https://dx.doi.org/10.13039/501100006041",
	"Max Planck Society":              "https://doi.org/10.13039/501100004189",
	"Ministry of Science, Research, and the Arts of the State of Baden-Württemberg (MWK), Juniorprofessor programme": "https://doi.org/10.13039/501100003542",
	"National Institute on Deafness and Other Communication Disorders":                                               "https://doi.org/10.13039/100000055",
	"NIH": "https://doi.org/10.13039/100000002",
	"NSF": "https://doi.org/10.13039/100000001",
	"Seventh Framework Programme (European Union Seventh Framework Programme)": "https://doi.org/10.13039/100011102",
	"The JPB Foundation": "https://doi.org/10.13039/100007457",
}

type Identifier struct {
	ID   string `xml:",chardata"`
	Type string `xml:"identifierType,attr"`
}

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

type FunderIdentifier struct {
	ID   string `xml:",chardata"`
	Type string `xml:"funderIdentifierType,attr"`
}

type FundingReference struct {
	Funder      string            `xml:"funderName"`
	AwardNumber string            `xml:"awardNumber"`
	Identifier  *FunderIdentifier `xml:"funderIdentifier"`
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
	XMLName        xml.Name `xml:"http://datacite.org/schema/kernel-4 resource"`
	Schema         string   `xml:"xmlns:xsi,attr"`
	SchemaLocation string   `xml:"xsi:schemaLocation,attr"`
	// Resource identifier (DOI)
	Identifier Identifier `xml:"identifier"`
	// Creators: Authors
	Creators     []Creator     `xml:"creators>creator"`
	Titles       []string      `xml:"titles>title"`
	Descriptions []Description `xml:"descriptions>description"`
	// RightsList: Licenses
	RightsList []Rights `xml:"rightsList>rights"`
	// Subjects: Keywords
	Subjects *[]string `xml:"subjects>subject,omitempty"`
	// RelatedIdentifiers: References
	RelatedIdentifiers []RelatedIdentifier `xml:"relatedIdentifiers>relatedIdentifier"`
	FundingReferences  *[]FundingReference `xml:"fundingReferences>fundingReference,omitempty"`
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
	// Size of the archive
	Sizes *[]string `xml:"sizes>size,omitempty"`
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
		XMLName:        xml.Name{Space: "http://datacite.org/schema/kernel-4", Local: "resource"},
		Schema:         Schema,
		SchemaLocation: SchemaLocation,
		Contributors:   []Contributor{Contributor{"German Neuroinformatics Node", "HostingInstitution"}},
		Publisher:      Publisher,
		Year:           time.Now().Year(),
		Dates:          []Date{Date{time.Now().Format("2006-01-02"), "Issued"}},
		Language:       Language,
		Version:        Version,
	}
}

func parseAuthorID(authorID string) *NameIdentifier {
	if authorID == "" {
		return nil
	}
	lowerID := strings.ToLower(authorID)
	if strings.HasPrefix(lowerID, "orcid") {
		// four blocks of four numbers separated by dash; last character can be X
		// https://support.orcid.org/hc/en-us/articles/360006897674-Structure-of-the-ORCID-Identifier
		var re = regexp.MustCompile(`([[:digit:]]{4}-){3}[[:digit:]]{3}[[:digit:]X]`)
		if orcid := re.Find([]byte(authorID)); orcid != nil {
			return &NameIdentifier{SchemeURI: "http://orcid.org/", Scheme: "ORCID", ID: string(orcid)}
		}
	} else if strings.HasPrefix(lowerID, "researcherid") {
		// couldn't find official description of format, but it seems to be:
		// letter, dash, four numbers, dash, four numbers
		var re = regexp.MustCompile(`[[:alpha:]](-[[:digit:]]{4}){2}`)
		if researcherid := re.Find([]byte(authorID)); researcherid != nil {
			// TODO: Find the proper values for these (publons.com?)
			return &NameIdentifier{SchemeURI: "http://publons.com/researcher/", Scheme: "ResercherID", ID: string(researcherid)}
		}
	}
	// unknown author ID type, or type identifier and format doesn't match regex: Return full string as ID
	return &NameIdentifier{SchemeURI: "", Scheme: "", ID: string(authorID)}
}

// FixSchemaAttrs adds the Schema and SchemaLocation attributes that can't be
// read from existing files when Unmarshaling.  See
// https://github.com/golang/go/issues/9519 for the issue with attributes that
// have a namespace prefix.
func (dc *DataCite) FixSchemaAttrs() {
	dc.Schema = Schema
	dc.SchemaLocation = SchemaLocation
}

func (dc *DataCite) AddAuthor(author *Author) {
	ident := parseAuthorID(author.ID)
	creator := Creator{
		Name:        fmt.Sprintf("%s, %s", author.LastName, author.FirstName),
		Identifier:  ident,
		Affiliation: author.Affiliation,
	}
	dc.Creators = append(dc.Creators, creator)
}

// AddAbstract is a convenience function for adding a Description with type
// "Abstract".
func (dc *DataCite) AddAbstract(abstract string) {
	dc.Descriptions = append(dc.Descriptions, Description{Content: abstract, Type: "Abstract"})
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
	if id, known := funderIDMap[funder]; known {
		fundref.Identifier = &FunderIdentifier{ID: id, Type: "Crossref Funder ID"}
	}
	if dc.FundingReferences == nil {
		dc.FundingReferences = &[]FundingReference{}
	}
	*dc.FundingReferences = append(*dc.FundingReferences, fundref)
}

// AddReference is a convenience function for appending a RelatedIdentifier
// that describes a referenced work. The RelatedIdentifier includes the
// identifier, relation type, and identifier type. A full citation string is
// also added to the Descriptions list.
func (dc *DataCite) AddReference(ref *Reference) {
	// Add info as RelatedIdentifier
	refIDParts := strings.SplitN(ref.ID, ":", 2)
	var relIDType, relID string
	if len(refIDParts) == 2 {
		relIDType = strings.TrimSpace(refIDParts[0])
		if ridt, ok := relIDTypeMap[strings.ToLower(relIDType)]; ok {
			relIDType = ridt
		}
		relID = strings.TrimSpace(refIDParts[1])
	} else {
		// No colon, add to ID as is
		relID = ref.ID
	}

	relatedIdentifier := RelatedIdentifier{Identifier: relID, Type: relIDType, RelationType: ref.RefType}
	dc.RelatedIdentifiers = append(dc.RelatedIdentifiers, relatedIdentifier)

	// Add citation string as Description
	var namecitation string
	if ref.Name != "" && ref.Citation != "" {
		namecitation = ref.Name + " " + ref.Citation
	} else {
		namecitation = ref.Name + ref.Citation
	}

	if !strings.HasSuffix(namecitation, ".") {
		namecitation += "."
	}
	refDesc := Description{Content: fmt.Sprintf("%s: %s (%s)", ref.RefType, namecitation, ref.GetURL()), Type: "Other"}

	dc.Descriptions = append(dc.Descriptions, refDesc)
}

func NewDataCiteFromYAML(info *RepositoryYAML) *DataCite {
	datacite := NewDataCite()
	for _, author := range info.Authors {
		datacite.AddAuthor(&author)
	}
	datacite.Titles = []string{info.Title}
	datacite.AddAbstract(info.Description)
	datacite.Subjects = &info.Keywords
	datacite.RightsList = []Rights{Rights{Name: info.License.Name, URL: info.License.URL}}
	for _, funding := range info.Funding {
		datacite.AddFunding(funding)
	}
	for _, ref := range info.References {
		datacite.AddReference(&ref)
	}
	datacite.SetResourceType(info.ResourceType)
	return &datacite
}

// UnmarshalFile reads an XML file specified by the given path and returns a
// populated DataCite struct.  Before returning, it adds the schema attributes
// for the top-level tag.  See also FixSchemaAttrs.
func UnmarshalFile(path string) (*DataCite, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	xmldata, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	dc := new(DataCite)
	err = xml.Unmarshal(xmldata, dc)

	if err != nil {
		return nil, err
	}
	dc.FixSchemaAttrs()
	return dc, nil
}

// AddURLs is a convenience function for appending three reference URLs:
// 1. The source repository URL;
// 2. The DOI fork repository URL;
// 3. The Archive URL.
// If the archive URL is valid and reachable, the Size of the archive is added
// as well.
func (dc *DataCite) AddURLs(repo, fork, archive string) {
	if repo != "" {
		relatedIdentifier := RelatedIdentifier{Identifier: repo, Type: "URL", RelationType: "IsVariantFormOf"}
		dc.RelatedIdentifiers = append(dc.RelatedIdentifiers, relatedIdentifier)
	}
	if fork != "" {
		relatedIdentifier := RelatedIdentifier{Identifier: fork, Type: "URL", RelationType: "IsVariantFormOf"}
		dc.RelatedIdentifiers = append(dc.RelatedIdentifiers, relatedIdentifier)
	}
	if archive != "" {
		relatedIdentifier := RelatedIdentifier{Identifier: archive, Type: "URL", RelationType: "IsVariantFormOf"}
		dc.RelatedIdentifiers = append(dc.RelatedIdentifiers, relatedIdentifier)
		if size, err := GetArchiveSize(archive); err == nil {
			dc.Sizes = &[]string{fmt.Sprintf("%d bytes", size)} // keep it in bytes so we can humanize it whenever we need to
		}
		// ignore error and don't add size
	}
}

// Marshal returns the marshalled version of the metadata structure, indented
// with tabs and with the appropriate XML header.
func (dc *DataCite) Marshal() (string, error) {
	dataciteXML, err := xml.MarshalIndent(dc, "", "  ")
	if err != nil {
		return "", err
	}

	return xml.Header + string(dataciteXML), nil
}
