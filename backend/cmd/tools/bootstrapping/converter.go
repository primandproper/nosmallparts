package main

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
)

const stageDirectionsCastID = "_STAGE_"
const stageDirectionsCastName = "Narrator"

// titleCase converts "OSTLER" to "Ostler", "FIRST CARRIER" to "First Carrier".
func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		words[i] = strings.ToUpper(string(w[0])) + strings.ToLower(w[1:])
	}
	return strings.Join(words, " ")
}

// bodyItem represents either a speech or stage direction in document order.
type bodyItem struct {
	ActNum   int
	Who      string
	LineText string
	IsStage  bool
}

// parseBodyItems extracts sp and stage elements from body InnerXML in document order.
// Returns items (with ActNum) and a map of who -> first speaker seen (for characters without persName).
func parseBodyItems(bodyXML string) ([]bodyItem, map[string]string, error) {
	dec := xml.NewDecoder(bytes.NewReader([]byte(bodyXML)))
	dec.Entity = xml.HTMLEntity

	var items []bodyItem
	speakerByWho := make(map[string]string)
	inSp := false
	inStage := false
	inAb := false
	inSpeaker := false
	currentAct := 1 // default to act 1 for content before first div1
	var currentWho string
	var textBuf, speakerBuf strings.Builder

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			local := t.Name.Local
			if local == "div1" {
				typ, n := "", ""
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "type":
						typ = a.Value
					case "n":
						n = a.Value
					}
				}
				switch typ {
				case "act":
					if num, err := strconv.Atoi(n); err == nil {
						currentAct = num
					}
				case "prologue", "induction", "dedication":
					currentAct = 0
				case "epilogue":
					currentAct = 99
				}
				continue
			}
			if local == "sp" {
				inSp = true
				inAb = false
				inSpeaker = false
				textBuf.Reset()
				speakerBuf.Reset()
				for _, a := range t.Attr {
					if a.Name.Local == "who" {
						currentWho = a.Value
						break
					}
				}
				continue
			}
			if local == "speaker" && inSp {
				inSpeaker = true
				speakerBuf.Reset()
				continue
			}
			if local == "ab" && inSp {
				inAb = true
				inSpeaker = false
				textBuf.Reset()
				continue
			}
			if local == "stage" {
				inStage = true
				textBuf.Reset()
				for _, a := range t.Attr {
					if a.Name.Local == "who" {
						currentWho = a.Value
						break
					}
				}
				continue
			}

		case xml.EndElement:
			if t.Name.Local == "speaker" && inSpeaker {
				// Normalize: collapse whitespace, trim. Folger uses ALL CAPS for speech prefixes;
				// title-case for display (OSTLER -> Ostler, FIRST CARRIER -> First Carrier)
				speakerText := strings.TrimSpace(speakerBuf.String())
				speakerText = strings.Join(strings.Fields(speakerText), " ")
				speakerText = titleCase(speakerText)
				who := strings.TrimPrefix(currentWho, "#")
				if idx := strings.Index(who, " "); idx > 0 {
					who = who[:idx]
				}
				if who != "" && speakerText != "" && speakerByWho[who] == "" {
					speakerByWho[who] = speakerText
				}
				inSpeaker = false
			}
			if t.Name.Local == "ab" && inSp && inAb {
				currentText := strings.TrimSpace(textBuf.String())
				currentText = strings.Join(strings.Fields(currentText), " ") // collapse newlines and extra spaces
				if currentText != "" {
					items = append(items, bodyItem{
						ActNum:   currentAct,
						IsStage:  false,
						Who:      currentWho,
						LineText: currentText,
					})
				}
				inAb = false
			}
			if t.Name.Local == "sp" && inSp {
				inSp = false
			}
			if t.Name.Local == "stage" && inStage {
				currentText := strings.TrimSpace(textBuf.String())
				currentText = strings.Join(strings.Fields(currentText), " ") // collapse newlines and extra spaces
				if currentText != "" {
					items = append(items, bodyItem{
						ActNum:   currentAct,
						IsStage:  true,
						Who:      currentWho,
						LineText: currentText,
					})
				}
				inStage = false
			}

		case xml.CharData:
			if inSp && inAb {
				textBuf.WriteString(string(t))
			} else if inStage {
				textBuf.WriteString(string(t))
			} else if inSpeaker {
				speakerBuf.WriteString(string(t))
			}
		}
	}

	return items, speakerByWho, nil
}

// actScript is a play script for a single act.
type actScript struct {
	ActNum int
	Script *playscripts.PlayScriptCreationRequestInput
}

// convertFolgerToPlayScript converts a parsed Folger TEI document into one PlayScriptCreationRequestInput per act.
func convertFolgerToPlayScript(tei *folgerTEI) ([]actScript, error) {
	people := collectPersonData(tei.Header.Profile.ParticDesc.ListPersons)
	people[stageDirectionsCastID] = personData{Name: stageDirectionsCastName}

	items, speakerByWho, err := parseBodyItems(tei.Text.Body.InnerXML)
	if err != nil {
		return nil, err
	}

	// Group items by act
	actItems := make(map[int][]bodyItem)
	for _, it := range items {
		actItems[it.ActNum] = append(actItems[it.ActNum], it)
	}

	title := strings.TrimSpace(tei.Header.FileDesc.TitleStmt.Title)
	author := strings.TrimSpace(tei.Header.FileDesc.TitleStmt.Author)

	var result []actScript
	for actNum := 0; actNum <= 99; actNum++ {
		actList, ok := actItems[actNum]
		if !ok || len(actList) == 0 {
			continue
		}

		// Build cast members from all lines (including narrator for stage directions)
		seen := make(map[string]bool)
		var castOrder []string
		for _, it := range actList {
			var ids []string
			if it.IsStage {
				ids = []string{stageDirectionsCastID}
			} else {
				ids = strings.Fields(strings.ReplaceAll(it.Who, "#", ""))
			}
			for _, id := range ids {
				if id != "" && !seen[id] {
					seen[id] = true
					castOrder = append(castOrder, id)
				}
			}
		}

		// Map Folger IDs to generated IDs for this act
		folgerToID := make(map[string]string)

		castMembers := make([]*playscripts.PlayScriptCastMemberCreationInput, 0, len(castOrder))
		for i, folgerID := range castOrder {
			generatedID := identifiers.New()
			folgerToID[folgerID] = generatedID

			pd := people[folgerID]
			name := pd.Name
			if name == "" {
				name = speakerByWho[folgerID]
			}
			if name == "" {
				name = folgerID
			}
			castMembers = append(castMembers, &playscripts.PlayScriptCastMemberCreationInput{
				ID:          generatedID,
				Name:        name,
				Description: pd.Description,
				SortOrder:   i,
			})
		}

		lines := make([]*playscripts.PlayScriptLineCreationRequestInput, 0, len(actList))
		for i, it := range actList {
			folgerID := stageDirectionsCastID
			if !it.IsStage {
				who := strings.TrimPrefix(it.Who, "#")
				if idx := strings.Index(who, " "); idx > 0 {
					folgerID = who[:idx]
				} else {
					folgerID = who
				}
				if folgerID == "" {
					folgerID = stageDirectionsCastID
				}
			}
			castMemberID := folgerToID[folgerID]

			lines = append(lines, &playscripts.PlayScriptLineCreationRequestInput{
				CastMemberID:  castMemberID,
				LineText:      it.LineText,
				Direction:     "",
				ScrollMode:    playscripts.DefaultScrollMode(),
				LineIndex:     i,
				AutoScrollWpm: 0,
			})
		}

		actTitle := title
		if actNum == 0 {
			actTitle = title + " (Prologue)"
		} else if actNum < 99 {
			actTitle = title + " (Act " + strconv.Itoa(actNum) + ")"
		} else {
			actTitle = title + " (Epilogue)"
		}

		result = append(result, actScript{
			ActNum: actNum,
			Script: &playscripts.PlayScriptCreationRequestInput{
				Title:       actTitle,
				Description: "Folger Shakespeare Library edition",
				Author:      author,
				CastMembers: castMembers,
				Lines:       lines,
				IsPublic:    true,
			},
		})
	}

	return result, nil
}
