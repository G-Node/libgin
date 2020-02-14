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
		Description{"This is something else\nSomething else entirely", "Other"}, // TODO: Line breaks
	}
	example.RightsList = Rights{"CC-BY", "http://creativecommons.org/licenses/by/4.0/"}
	example.Subjects = []string{"One", "Two", "Three"}
	example.FundingReferences = []FundingReference{
		FundingReference{"DFG", "DFG.12345", "Crossref Funder ID"},
		FundingReference{"EU", "EU.12345", "Crossref Funder ID"},
	}
	// example.AddFunding("DFG, DFG.12345")
	// example.AddFunding("EU, EU.12345")
	example.RelatedIdentifiers = []RelatedIdentifier{
		RelatedIdentifier{"10.1111/fake.doi", "DOI", "IsDescribedBy"},
		RelatedIdentifier{"10.2222/fake.doi", "DOI", "IsSupplementTo"},
		RelatedIdentifier{"10.3333/fake.doi", "DOI", "IsReferencedBy"},
	}
	example.SetResourceType("Dataset")

	dataciteXML, err := xml.MarshalIndent(example, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal: %v\n", err)
	}

	fmt.Println(xml.Header + string(dataciteXML))
}
