package libgin

import (
	"encoding/xml"
	"fmt"
	"testing"
)

func Test_DataCiteMarshal(t *testing.T) {
	example := NewDataCite()
	example.Creators = []Creator{
		Creator{"Achilleas", &NameIdentifier{"0010", "http://orcid.org", "ORCID"}, "University of Bob"},
		Creator{"Bob", &NameIdentifier{"1111", "http://orcid.org", "ORCID"}, "University of University"},
		Creator{"Orphan", nil, ""},
	}

	example.Titles = []string{"This is a sample"}
	example.AddAbstract("This is the abstract")
	example.RightsList = []Rights{Rights{"CC-BY", "http://creativecommons.org/licenses/by/4.0/"}}
	example.Subjects = []string{"One", "Two", "Three"}
	example.AddFunding("DFG, DFG.12345")
	example.AddFunding("EU, EU.12345")
	example.SetResourceType("Dataset")

	example.AddReference(&Reference{ID: "doi:10.1111/example.doi", Reftype: "IsDescribedBy", Name: "Manuscript title for reference."})
	example.AddReference(&Reference{ID: "arxiv:10.2222/example.doi", Reftype: "IsSupplementTo", Name: "Some other work"})
	example.AddReference(&Reference{ID: "doi:10.3333/example.doi", Reftype: "IsReferencedBy", Name: "A work that references this dataset."})

	_, err := xml.MarshalIndent(example, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal: %v\n", err)
	}

	// fmt.Println(xml.Header + string(dataciteXML))
}

func Test_DataCiteFromRegInfo(t *testing.T) {
	authors := []Author{
		Author{
			FirstName:   "GivenName1",
			LastName:    "FamilyName1",
			Affiliation: "Affiliation1",
			ID:          "ORCID:0000-0001-2345-6789",
		},
		Author{
			FirstName:   "GivenName2",
			LastName:    "FamilyName2",
			Affiliation: "Affiliation2",
			ID:          "ResearcherID:X-1234-5678",
		},
		Author{
			FirstName: "GivenName3",
			LastName:  "FamilyName3",
		},
	}
	references := []Reference{
		Reference{
			ID:      "doi:10.xxx/zzzz",
			Reftype: "IsSupplementTo",
			Name:    "PublicationName1",
		},
		Reference{
			ID:      "arxiv:mmmm.nnnn",
			Reftype: "IsDescribedBy",
			Name:    "PublicationName2",
		},
		Reference{
			ID:      "pmid:nnnnnnnn",
			Reftype: "IsReferencedBy",
			Name:    "PublicationName3",
		},
	}
	regInfo := &DOIRegInfo{
		Authors:      authors,
		Title:        "This is a sample",
		Description:  "This is the abstract",
		Keywords:     []string{"Neuroscience", "Electrophysiology"},
		License:      &License{"Creative Commons CC0 1.0 Public Domain Dedication", "https://creativecommons.org/publicdomain/zero/1.0/"},
		Funding:      []string{"DFG, DFG.12345", "EU, EU.12345"},
		References:   references,
		ResourceType: "Dataset",
	}
	example := NewDataCiteFromRegInfo(regInfo)
	dataciteXML, err := xml.MarshalIndent(example, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal: %v\n", err)
	}

	fmt.Println(xml.Header + string(dataciteXML))
}
