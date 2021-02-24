package libgin

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// RepositoryYAML is used to read the information provided by a GIN user
// through the datacite.yml file. This data is usually used to populate the
// DataCite and RepositoryMetadata types.
type RepositoryYAML struct {
	Authors         []Author    `yaml:"authors"`
	Title           string      `yaml:"title"`
	Description     string      `yaml:"description"`
	Keywords        []string    `yaml:"keywords"`
	License         *License    `yaml:"license,omitempty"`
	Funding         []string    `yaml:"funding,omitempty"`
	References      []Reference `yaml:"references,omitempty"`
	TemplateVersion string      `yaml:"templateversion,omitempty"`
	ResourceType    string      `yaml:"resourcetype"`
}

type Author struct {
	FirstName   string `yaml:"firstname"`
	LastName    string `yaml:"lastname"`
	Affiliation string `yaml:"affiliation,omitempty"`
	ID          string `yaml:"id,omitempty"`
}

type License struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Reference struct {
	ID       string `yaml:"id,omitempty"`
	RefType  string `yaml:"reftype,omitempty"`
	Name     string `yaml:"name,omitempty"`     // deprecated, but still read for older versions
	Citation string `yaml:"citation,omitempty"` // meant to replace Name
}

type GINUser struct {
	Username string
	Email    string
	RealName string
}

// RepositoryMetadata can contain all known metadata for a registered (or
// to-be-registered) repository. To do this, it embeds the
type RepositoryMetadata struct {
	// YAMLData is the original data coming from the repository
	YAMLData *RepositoryYAML
	// DataCite is the struct that produces the XML file
	*DataCite
	// The following are computed or generated from external info and don't
	// all show up in the YAML or XML files

	// The user that sent the request
	RequestingUser *GINUser
	// Should be full repository path (<user>/<reponame>)
	SourceRepository string
	// Should be full repository path of the snapshot fork (doi/<rpeoname>)
	ForkRepository string
	// UUID calculated from unique repository path or randomly assigned
	UUID string
}

// NOTE: TEMPORARY COPIES FROM gin-doi

// UUIDMap is a map between registered repositories and their UUIDs for
// datasets registered before the new UUID generation method was implemented.
// This map is required because the current method of computing UUIDs differs
// from the older method and this lookup is used to handle the old-method
// UUIDs.
var UUIDMap = map[string]string{
	"INT/multielectrode_grasp":                   "f83565d148510fede8a277f660e1a419",
	"ajkumaraswamy/HB-PAC_disinhibitory_network": "1090f803258557299d287c4d44a541b2",
	"steffi/Kleineidam_et_al_2017":               "f53069de4c4921a3cfa8f17d55ef98bb",
	"Churan/Morris_et_al_Frontiers_2016":         "97bc1456d3f4bca2d945357b3ec92029",
	"fabee/efish_locking":                        "6953bbf0087ba444b2d549b759de4a06",
}

// RepoPathToUUID computes a UUID from a repository path.
func RepoPathToUUID(URI string) string {
	if doi, ok := UUIDMap[URI]; ok {
		return doi
	}
	currMd5 := md5.Sum([]byte(URI))
	return hex.EncodeToString(currMd5[:])
}

// DOIRequestData is used to transmit data from GIN to DOI when a registration
// request is triggered.
type DOIRequestData struct {
	Username   string
	Realname   string
	Repository string
	Email      string
}

// DOIRegInfo holds all the metadata and information necessary for a DOI registration request.
// Deprecated and obsolete: Marked for removal
type DOIRegInfo struct {
	Missing         []string
	DOI             string
	UUID            string
	FileName        string
	FileSize        string
	Title           string
	Authors         []Author
	Description     string
	Keywords        []string
	References      []Reference
	Funding         []string
	License         *License
	ResourceType    string
	DateTime        time.Time
	TemplateVersion string
}

func (c *DOIRegInfo) GetType() string {
	if c.ResourceType != "" {
		return c.ResourceType
	}
	return "Dataset"
}

func (c *DOIRegInfo) GetCitation() string {
	var authors string
	for _, auth := range c.Authors {
		if len(auth.FirstName) > 0 {
			authors += fmt.Sprintf("%s %s, ", auth.LastName, string(auth.FirstName[0]))
		} else {
			authors += fmt.Sprintf("%s, ", auth.LastName)
		}
	}
	return fmt.Sprintf("%s (%s) %s. G-Node. https://doi.org/%s", authors, c.Year(), c.Title, c.DOI)
}

func (c *DOIRegInfo) Year() string {
	return fmt.Sprintf("%d", c.DateTime.Year())
}

func (c *DOIRegInfo) ISODate() string {
	return c.DateTime.Format("2006-01-02")
}

func PrettyDate(dt *time.Time) string {
	return dt.Format("02 Jan. 2006")
}

func (c *Author) GetValidID() *NamedIdentifier {
	if c.ID == "" {
		return nil
	}
	if strings.HasPrefix(strings.ToLower(c.ID), "orcid") {
		// four blocks of four numbers separated by dash; last character can be X
		// https://support.orcid.org/hc/en-us/articles/360006897674-Structure-of-the-ORCID-Identifier
		var re = regexp.MustCompile(`([[:digit:]]{4}-){3}[[:digit:]]{3}[[:digit:]X]`)
		if orcid := re.Find([]byte(c.ID)); orcid != nil {
			return &NamedIdentifier{SchemeURI: "http://orcid.org/", Scheme: "ORCID", ID: string(orcid)}
		}
	}
	return nil
}
func (a *Author) RenderAuthor() string {
	auth := fmt.Sprintf("%s,%s;%s;%s", a.LastName, a.FirstName, a.Affiliation, a.ID)
	return strings.Replace(strings.TrimRight(auth, ";"), ";;", ";", -1)
}

type NamedIdentifier struct {
	SchemeURI string
	Scheme    string
	ID        string
}

func (ref Reference) GetURL() string {
	idparts := strings.SplitN(ref.ID, ":", 2)
	if len(idparts) != 2 {
		// Malformed ID (no colon)
		return ""
	}
	source := idparts[0]
	idnum := idparts[1]

	var prefix string
	switch strings.ToLower(source) {
	case "doi":
		prefix = "https://doi.org/"
	case "arxiv":
		// https://arxiv.org/help/arxiv_identifier_for_services
		prefix = "https://arxiv.org/abs/"
	case "pmid":
		// https://www.ncbi.nlm.nih.gov/books/NBK3862/#linkshelp.Retrieve_PubMed_Citations
		prefix = "https://www.ncbi.nlm.nih.gov/pubmed/"
	case "url":
		// simple URL: no prefix
		prefix = ""
	default:
		// Return an empty string to make the reflink inactive
		return ""
	}

	return fmt.Sprintf("%s%s", prefix, idnum)
}

func IsRegisteredDOI(doi string) bool {
	url := fmt.Sprintf("https://doi.org/%s", doi)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Could not query for DOI: %s at %s", doi, url)
		return false
	}
	if resp.StatusCode != http.StatusNotFound {
		return true
	}
	return false
}
