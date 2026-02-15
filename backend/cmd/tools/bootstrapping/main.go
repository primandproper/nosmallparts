package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// sanitizeForFilename converts a title like "Henry IV, Part I" to "Henry_IV_Part_I".
func sanitizeForFilename(s string) string {
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`[^\p{L}\p{N}\s-]`).ReplaceAllString(s, "") // remove punctuation except hyphen
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	return s
}

func main() {
	inputDir := flag.String("input", "ShakespearFolger", "Directory containing Folger XML files")
	outputDir := flag.String("output", "output", "Directory to write JSON PlayScript files")
	clean := flag.Bool("clean", false, "Delete existing output directory before generating")
	flag.Parse()

	if *clean {
		if err := os.RemoveAll(*outputDir); err != nil {
			log.Fatalf("Failed to remove output directory %s: %v", *outputDir, err)
		}
	}

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	entries, err := os.ReadDir(*inputDir)
	if err != nil {
		log.Fatalf("Failed to read input directory %s: %v", *inputDir, err)
	}

	// Sort XML files for deterministic play indexing
	var xmlFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".xml" {
			xmlFiles = append(xmlFiles, e.Name())
		}
	}
	sort.Strings(xmlFiles)

	converted := 0
	playIndex := 1
	for _, name := range xmlFiles {
		inputPath := filepath.Join(*inputDir, name)
		data, err := os.ReadFile(inputPath)
		if err != nil {
			log.Printf("Failed to read %s: %v", inputPath, err)
			continue
		}

		var tei folgerTEI
		if err := xml.Unmarshal(data, &tei); err != nil {
			log.Printf("Failed to parse XML %s: %v", inputPath, err)
			continue
		}

		actScripts, err := convertFolgerToPlayScript(&tei)
		if err != nil {
			log.Printf("Failed to convert %s: %v", inputPath, err)
			continue
		}

		if len(actScripts) == 0 {
			continue
		}

		title := strings.TrimSpace(tei.Header.FileDesc.TitleStmt.Title)
		sanitizedTitle := sanitizeForFilename(title)

		for _, as := range actScripts {
			var actLabel string
			switch as.ActNum {
			case 0:
				actLabel = "Prologue"
			case 99:
				actLabel = "Epilogue"
			default:
				actLabel = fmt.Sprintf("Act_%d", as.ActNum)
			}

			outputName := fmt.Sprintf("%03d_%s_%s.json", playIndex, sanitizedTitle, actLabel)
			outputPath := filepath.Join(*outputDir, outputName)

			jsonBytes, err := json.MarshalIndent(as.Script, "", "  ")
			if err != nil {
				log.Printf("Failed to marshal JSON for %s: %v", outputName, err)
				continue
			}

			if err := os.WriteFile(outputPath, jsonBytes, 0o644); err != nil {
				log.Printf("Failed to write %s: %v", outputPath, err)
				continue
			}

			converted++
			fmt.Printf("Converted %s -> %s (%d lines)\n", name, outputName, len(as.Script.Lines))
		}

		playIndex++
	}

	fmt.Printf("\nDone. Converted %d act files to JSON in %s/\n", converted, *outputDir)
}
