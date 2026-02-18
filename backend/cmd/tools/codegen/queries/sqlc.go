package main

import (
	"fmt"
	"strings"
)

type QueryType string

const (
	ExecType     QueryType = ":exec"
	ExecRowsType QueryType = ":execrows"
	ManyType     QueryType = ":many"
	OneType      QueryType = ":one"
)

type QueryAnnotation struct {
	Name string
	Type QueryType
}

type Query struct {
	Content    string
	Annotation QueryAnnotation
}

func (q *Query) Render() string {
	content := q.Content
	if !strings.HasSuffix(content, ";") {
		content += ";"
	}

	return fmt.Sprintf("-- name: %s %s\n%s\n", q.Annotation.Name, q.Annotation.Type, content)
}
