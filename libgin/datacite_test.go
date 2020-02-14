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
	}
	example.RightsList = Rights{"CC-BY", "http://creativecommons.org/licenses/by/4.0/"}
	example.Subjects = []string{"One", "Two", "Three"}
	example.AddFunding("DFG, DFG.12345")
	example.AddFunding("EU, EU.12345")
	example.SetResourceType("Dataset")

	example.AddReference(&Reference{ID: "doi:10.1111/example.doi", Reftype: "IsDescribedBy", Name: "Manuscript title for reference."})
	example.AddReference(&Reference{ID: "arxiv:10.2222/example.doi", Reftype: "IsSupplementTo", Name: "Some other work"})
	example.AddReference(&Reference{ID: "doi:10.3333/example.doi", Reftype: "IsReferencedBy", Name: "A work that references this dataset."})

	dataciteXML, err := xml.MarshalIndent(example, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal: %v\n", err)
	}

	fmt.Println(xml.Header + string(dataciteXML))
}
