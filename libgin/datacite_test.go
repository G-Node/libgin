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
	example.Descriptions = []Description{
		Description{"This is the abstract", "Abstract"},
		Description{"IsDescribedBy: Manuscript title for reference. doi:10.1111/example.doi", "Other"}, // TODO: Line breaks
		Description{"IsSupplementTo: Some other work. arxiv:10.2222/example.doi", "Other"},
		Description{"IsReferencedBy: A work that references this dataset. doi:10.3333/example.doi", "Other"},
	}
	example.RightsList = Rights{"CC-BY", "http://creativecommons.org/licenses/by/4.0/"}
	example.Subjects = []string{"One", "Two", "Three"}
	example.AddFunding("DFG, DFG.12345")
	example.AddFunding("EU, EU.12345")
	example.RelatedIdentifiers = []RelatedIdentifier{
		RelatedIdentifier{"10.1111/example.doi", "DOI", "IsDescribedBy"},
		RelatedIdentifier{"10.2222/example.doi", "arxiv", "IsSupplementTo"},
		RelatedIdentifier{"10.3333/example.doi", "DOI", "IsReferencedBy"},
	}
	example.SetResourceType("Dataset")

	dataciteXML, err := xml.MarshalIndent(example, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal: %v\n", err)
	}

	fmt.Println(xml.Header + string(dataciteXML))
}
