package main

import (
	"encoding/xml"
	"regexp"
	"strings"
)

// folgerTEI represents the root TEI element from Folger Shakespeare XML.
type folgerTEI struct {
	XMLName xml.Name     `xml:"http://www.tei-c.org/ns/1.0 TEI"`
	Text    folgerText   `xml:"http://www.tei-c.org/ns/1.0 text"`
	Header  folgerHeader `xml:"http://www.tei-c.org/ns/1.0 teiHeader"`
}

type folgerHeader struct {
	FileDesc folgerFileDesc `xml:"http://www.tei-c.org/ns/1.0 fileDesc"`
	Profile  folgerProfile  `xml:"http://www.tei-c.org/ns/1.0 profileDesc"`
}

type folgerFileDesc struct {
	TitleStmt folgerTitleStmt `xml:"http://www.tei-c.org/ns/1.0 titleStmt"`
}

type folgerTitleStmt struct {
	Title  string `xml:"http://www.tei-c.org/ns/1.0 title"`
	Author string `xml:"http://www.tei-c.org/ns/1.0 author"`
}

type folgerProfile struct {
	ParticDesc folgerParticDesc `xml:"http://www.tei-c.org/ns/1.0 particDesc"`
}

type folgerParticDesc struct {
	ListPersons []folgerListPerson `xml:"http://www.tei-c.org/ns/1.0 listPerson"`
}

type folgerListPerson struct {
	Head      string             `xml:"http://www.tei-c.org/ns/1.0 head"`
	Person    []folgerPerson     `xml:"http://www.tei-c.org/ns/1.0 person"`
	PersonGrp []folgerPersonGrp  `xml:"http://www.tei-c.org/ns/1.0 personGrp"`
	Nested    []folgerListPerson `xml:"http://www.tei-c.org/ns/1.0 listPerson"`
}

type folgerPerson struct {
	ID       string         `xml:"id,attr"`
	PersName folgerPersName `xml:"http://www.tei-c.org/ns/1.0 persName"`
	State    folgerState    `xml:"http://www.tei-c.org/ns/1.0 state"`
}

type folgerPersonGrp struct {
	ID string `xml:"id,attr"`
	P  string `xml:"http://www.tei-c.org/ns/1.0 p"`
}

type folgerPersName struct {
	InnerXML string `xml:",innerxml"`
	// Use local name only - nested elements inherit default TEI namespace from root
	Name     string `xml:"name"`
	RoleName string `xml:"roleName"`
}

type folgerState struct {
	P struct {
		InnerXML string `xml:",innerxml"`
	} `xml:"p"`
}

type folgerText struct {
	Body folgerBody `xml:"http://www.tei-c.org/ns/1.0 body"`
}

type folgerBody struct {
	InnerXML string `xml:",innerxml"`
}

// extractTextFromTEIContent extracts visible text from TEI InnerXML (w, c, pc elements).
// It removes milestone, lb, and other structural elements, keeping only character content.
func extractTextFromTEIContent(inner string) string {
	// Match content of w, c, and pc elements. These contain the actual text.
	// Format: <w ...>text</w>, <c ...> </c>, <pc ...>.</pc>
	re := regexp.MustCompile(`<(?:w|c|pc)[^>]*>([^<]*)</(?:w|c|pc)>`)
	matches := re.FindAllStringSubmatch(inner, -1)
	var sb strings.Builder
	for _, m := range matches {
		if len(m) > 1 {
			sb.WriteString(m[1])
		}
	}
	return strings.TrimSpace(sb.String())
}

// personName extracts display name from persName - prefers Name or RoleName, else InnerXML.
func (p folgerPersName) personName() string {
	if p.Name != "" {
		return strings.TrimSpace(p.Name)
	}
	if p.RoleName != "" {
		return strings.TrimSpace(p.RoleName)
	}
	// Fallback: strip tags from InnerXML to get text
	return strings.TrimSpace(stripXMLTags(p.InnerXML))
}

func stripXMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(s, "")
}

// extractStateDescription extracts full text from state's <p> content, including nested
// elements like <name>, and replaces TEI break elements (lb, pb) with a space.
func extractStateDescription(st folgerState) string {
	inner := st.P.InnerXML
	if inner == "" {
		return ""
	}
	// Replace line/page break elements with space so we don't get "Illyria(later"
	inner = regexp.MustCompile(`<(?:lb|pb)[^>]*/?>`).ReplaceAllString(inner, " ")
	return strings.TrimSpace(collapseSpace(stripXMLTags(inner)))
}

func collapseSpace(s string) string {
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(s, " "))
}

// personData holds name and description for a cast member.
type personData struct {
	Name        string
	Description string
}

// collectPersonData recursively extracts person ID -> {name, description} from nested listPerson.
func collectPersonData(lists []folgerListPerson) map[string]personData {
	result := make(map[string]personData)
	var visit func([]folgerListPerson)
	visit = func(lists []folgerListPerson) {
		for _, list := range lists {
			for _, p := range list.Person {
				if p.ID != "" {
					name := p.PersName.personName()
					desc := extractStateDescription(p.State)
					if name != "" {
						result[p.ID] = personData{Name: name, Description: desc}
					}
					// When persName is empty, leave unset so converter can use speaker as fallback
				}
			}
			for _, grp := range list.PersonGrp {
				if grp.ID != "" {
					name := strings.TrimSpace(stripXMLTags(grp.P))
					if name == "" {
						name = grp.ID
					}
					result[grp.ID] = personData{Name: name}
				}
			}
			visit(list.Nested)
		}
	}
	visit(lists)
	return result
}
