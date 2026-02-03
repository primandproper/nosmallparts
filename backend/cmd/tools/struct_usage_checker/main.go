package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
)

const (
	// PatternTypeFunctionCallArgs checks struct usage in function call arguments (e.g., querier.Method(ctx, db, &Params{})).
	PatternTypeFunctionCallArgs = "function_call_args"
	// PatternTypeStructLiterals checks struct usage in struct literals (e.g., return &SomeType{}).
	PatternTypeStructLiterals = "struct_literals"
)

type CheckConfig struct {
	TypeFilter      func(string) bool
	SourcePkg       string
	PatternType     string
	TargetFieldName string
	Description     string
	TargetPkgs      []string
}

func noTestFiles(f os.FileInfo) bool {
	return !f.IsDir() && !strings.HasSuffix(f.Name(), "_test.go")
}

func fetchTypesForPackage(pkg string, nameFilter func(string) bool) map[string]*ast.StructType {
	fileset := token.NewFileSet()
	astPkg, err := parser.ParseDir(fileset, pkg, noTestFiles, parser.AllErrors)
	if err != nil {
		log.Fatalf("failed to parse package: %v", err)
	}

	if len(astPkg) != 1 {
		return nil
	}

	foundTypes := map[string]*ast.StructType{}
	for _, p := range astPkg {
		for _, f := range p.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				switch t := n.(type) {
				case *ast.TypeSpec:
					if structType, ok := t.Type.(*ast.StructType); ok && nameFilter(t.Name.Name) {
						foundTypes[t.Name.Name] = structType
					}
				}
				return true
			})
		}
	}

	return foundTypes
}

func getFieldsForStruct(structType *ast.StructType) map[string]string {
	structFields := make(map[string]string)

	for _, field := range structType.Fields.List {
		fieldType := ""
		switch t := field.Type.(type) {
		case *ast.Ident:
			fieldType = t.Name
		case *ast.SelectorExpr:
			fieldType = fmt.Sprintf("%s.%s", t.X, t.Sel.Name)
		}

		for _, name := range field.Names {
			if name.Name != "_" {
				structFields[name.Name] = fieldType
			}
		}
	}

	return structFields
}

// getSourceImportPath converts a relative package path to a full import path.
// Assumes the module is github.com/verygoodsoftwarenotvirus/nosmallparts/backend.
func getSourceImportPath(relativePath string) string {
	return "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/" + relativePath
}

// buildImportMap builds a map of import aliases to their package paths for a file.
func buildImportMap(f *ast.File) map[string]string {
	importMap := make(map[string]string)
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			// Extract package name from import path
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}
		importMap[alias] = importPath
	}
	return importMap
}

func comparePackages(config *CheckConfig, auxPackage string) (int, error) {
	paramTypes := fetchTypesForPackage(config.SourcePkg, config.TypeFilter)
	sourceImportPath := getSourceImportPath(config.SourcePkg)

	fileset := token.NewFileSet()
	astPkg, err := parser.ParseDir(fileset, auxPackage, noTestFiles, parser.AllErrors)
	if err != nil {
		log.Fatalf("failed to parse package: %p", err)
	}

	errors := &multierror.Error{}
	structCount := 0

	var count int
	for _, p := range astPkg {
		for fileName, f := range p.Files {
			importMap := buildImportMap(f)
			ast.Inspect(f, func(n ast.Node) bool {
				switch config.PatternType {
				case PatternTypeFunctionCallArgs:
					if et, ok := n.(*ast.CallExpr); ok {
						count, err = checkFunctionCallArgs(et, paramTypes, fileName, config.TargetFieldName, fileset, sourceImportPath, importMap)
						structCount += count
						if err != nil {
							errors = multierror.Append(errors, err)
						}
					}
				case PatternTypeStructLiterals:
					if cl, ok := n.(*ast.CompositeLit); ok {
						count, err = checkStructLiteral(cl, paramTypes, fileName, fileset, sourceImportPath, importMap)
						structCount += count
						if err != nil {
							errors = multierror.Append(errors, err)
						}
					}
				default:
					// Check both patterns for backward compatibility
					if et, ok := n.(*ast.CallExpr); ok {
						count, err = checkFunctionCallArgs(et, paramTypes, fileName, config.TargetFieldName, fileset, sourceImportPath, importMap)
						structCount += count
						if err != nil {
							errors = multierror.Append(errors, err)
						}
					}
					if cl, ok := n.(*ast.CompositeLit); ok {
						count, err = checkStructLiteral(cl, paramTypes, fileName, fileset, sourceImportPath, importMap)
						structCount += count
						if err != nil {
							errors = multierror.Append(errors, err)
						}
					}
				}
				return true
			})
		}
	}

	return structCount, errors.ErrorOrNil()
}

func checkFunctionCallArgs(et *ast.CallExpr, paramTypes map[string]*ast.StructType, fileName, targetFieldName string, fileset *token.FileSet, sourceImportPath string, importMap map[string]string) (int, error) {
	if targetFieldName == "" {
		return 0, nil // Skip if no target field name specified
	}

	switch ft := et.Fun.(type) {
	case *ast.SelectorExpr:
		if ftIdent, ok := ft.X.(*ast.SelectorExpr); ok {
			if ftIdent.Sel.Name == targetFieldName {
				if len(et.Args) == 3 {
					thirdParam := et.Args[2]
					if ue, isUE := thirdParam.(*ast.UnaryExpr); isUE {
						if se, isSE := ue.X.(*ast.CompositeLit); isSE {
							return checkCompositeFields(se, paramTypes, fileName, fileset, sourceImportPath, importMap)
						}
					}
				}
			}
		}
	}
	return 0, nil
}

func checkStructLiteral(cl *ast.CompositeLit, paramTypes map[string]*ast.StructType, fileName string, fileset *token.FileSet, sourceImportPath string, importMap map[string]string) (int, error) {
	return checkCompositeFields(cl, paramTypes, fileName, fileset, sourceImportPath, importMap)
}

func checkCompositeFields(se *ast.CompositeLit, paramTypes map[string]*ast.StructType, fileName string, fileset *token.FileSet, sourceImportPath string, importMap map[string]string) (int, error) {
	var errors *multierror.Error
	pos := fileset.Position(se.Pos())
	lineNum := pos.Line
	structCount := 0

	if tt, isType := se.Type.(*ast.SelectorExpr); isType {
		// Check if the package selector refers to the source package
		if pkgIdent, ok := tt.X.(*ast.Ident); ok {
			importPath, found := importMap[pkgIdent.Name]
			if !found {
				// If not in import map, it might be a built-in or local type, skip it
				return 0, nil
			}
			// Compare the full import path with the source import path
			if importPath != sourceImportPath {
				// This struct is from a different package, skip it
				return 0, nil
			}
		}

		lookup := tt.Sel.Name
		if fieldDef, present := paramTypes[lookup]; present {
			structCount = 1 // We're checking this struct
			fieldsUsed := map[string]string{}
			for _, el := range se.Elts {
				if kv, isKV := el.(*ast.KeyValueExpr); isKV {
					if ident, isIdent := kv.Key.(*ast.Ident); isIdent {
						fieldsUsed[ident.Name] = ""
					}
				}
			}

			structsForField := getFieldsForStruct(fieldDef)
			for fieldName := range structsForField {
				if _, used := fieldsUsed[fieldName]; !used {
					errors = multierror.Append(errors, fmt.Errorf("field %s not used in %s in backend/%s:%d", fieldName, lookup, fileName, lineNum))
				}
			}
		}
	}
	// Also handle direct type references like &SomeType{} without package selector
	// These are from the current package, so we check them if the current package matches
	if ident, isIdent := se.Type.(*ast.Ident); isIdent {
		lookup := ident.Name
		if fieldDef, present := paramTypes[lookup]; present {
			structCount = 1 // We're checking this struct
			fieldsUsed := map[string]string{}
			for _, el := range se.Elts {
				if kv, isKV := el.(*ast.KeyValueExpr); isKV {
					var identKey *ast.Ident
					if identKey, isIdent = kv.Key.(*ast.Ident); isIdent {
						fieldsUsed[identKey.Name] = ""
					}
				}
			}

			structsForField := getFieldsForStruct(fieldDef)
			for fieldName := range structsForField {
				if _, used := fieldsUsed[fieldName]; !used {
					errors = multierror.Append(errors, fmt.Errorf("field %s not used in %s in %s:%d", fieldName, lookup, fileName, lineNum))
				}
			}
		}
	}

	return structCount, errors.ErrorOrNil()
}

func main() {
	var errors *multierror.Error

	// Configuration for different check types
	configs := []*CheckConfig{
		// Postgres generated packages - check *Params structs in function calls
		{
			SourcePkg:       "internal/repositories/postgres/auditlogentries/generated",
			TargetPkgs:      []string{"internal/repositories/postgres/auditlogentries"},
			TypeFilter:      func(s string) bool { return strings.HasSuffix(s, "Params") },
			PatternType:     PatternTypeFunctionCallArgs,
			TargetFieldName: "generatedQuerier",
			Description:     "auditlogentries Params",
		},
		{
			SourcePkg:       "internal/repositories/postgres/auth/generated",
			TargetPkgs:      []string{"internal/repositories/postgres/auth"},
			TypeFilter:      func(s string) bool { return strings.HasSuffix(s, "Params") },
			PatternType:     PatternTypeFunctionCallArgs,
			TargetFieldName: "generatedQuerier",
			Description:     "auth Params",
		},
		{
			SourcePkg:       "internal/repositories/postgres/identity/generated",
			TargetPkgs:      []string{"internal/repositories/postgres/identity"},
			TypeFilter:      func(s string) bool { return strings.HasSuffix(s, "Params") },
			PatternType:     PatternTypeFunctionCallArgs,
			TargetFieldName: "generatedQuerier",
			Description:     "identity Params",
		},
		{
			SourcePkg:       "internal/repositories/postgres/maintenance/generated",
			TargetPkgs:      []string{"internal/repositories/postgres/maintenance"},
			TypeFilter:      func(s string) bool { return strings.HasSuffix(s, "Params") },
			PatternType:     PatternTypeFunctionCallArgs,
			TargetFieldName: "generatedQuerier",
			Description:     "maintenance Params",
		},
		{
			SourcePkg:       "internal/repositories/postgres/notifications/generated",
			TargetPkgs:      []string{"internal/repositories/postgres/notifications"},
			TypeFilter:      func(s string) bool { return strings.HasSuffix(s, "Params") },
			PatternType:     PatternTypeFunctionCallArgs,
			TargetFieldName: "generatedQuerier",
			Description:     "notifications Params",
		},
		{
			SourcePkg:       "internal/repositories/postgres/oauth/generated",
			TargetPkgs:      []string{"internal/repositories/postgres/oauth"},
			TypeFilter:      func(s string) bool { return strings.HasSuffix(s, "Params") },
			PatternType:     PatternTypeFunctionCallArgs,
			TargetFieldName: "generatedQuerier",
			Description:     "oauth Params",
		},
		{
			SourcePkg:       "internal/repositories/postgres/settings/generated",
			TargetPkgs:      []string{"internal/repositories/postgres/settings"},
			TypeFilter:      func(s string) bool { return strings.HasSuffix(s, "Params") },
			PatternType:     PatternTypeFunctionCallArgs,
			TargetFieldName: "generatedQuerier",
			Description:     "settings Params",
		},
		{
			SourcePkg:       "internal/repositories/postgres/webhooks/generated",
			TargetPkgs:      []string{"internal/repositories/postgres/webhooks"},
			TypeFilter:      func(s string) bool { return strings.HasSuffix(s, "Params") },
			PatternType:     PatternTypeFunctionCallArgs,
			TargetFieldName: "generatedQuerier",
			Description:     "webhooks Params",
		},
		// gRPC service converter packages - check struct literals in return/assignments
		{
			SourcePkg:   "internal/domain/identity",
			TargetPkgs:  []string{"internal/services/identity/grpc/converters"},
			TypeFilter:  func(s string) bool { return strings.Contains(s, "Input") || strings.Contains(s, "Response") },
			PatternType: PatternTypeStructLiterals,
			Description: "identity gRPC converters",
		},
		{
			SourcePkg:   "internal/domain/oauth",
			TargetPkgs:  []string{"internal/services/oauth/grpc/converters"},
			TypeFilter:  func(s string) bool { return strings.Contains(s, "Input") || strings.Contains(s, "Response") },
			PatternType: PatternTypeStructLiterals,
			Description: "oauth gRPC converters",
		},
		{
			SourcePkg:   "internal/domain/webhooks",
			TargetPkgs:  []string{"internal/services/webhooks/grpc/converters"},
			TypeFilter:  func(s string) bool { return strings.Contains(s, "Input") || strings.Contains(s, "Response") },
			PatternType: PatternTypeStructLiterals,
			Description: "webhooks gRPC converters",
		},
		{
			SourcePkg:   "internal/domain/settings",
			TargetPkgs:  []string{"internal/services/settings/grpc/converters"},
			TypeFilter:  func(s string) bool { return strings.Contains(s, "Input") || strings.Contains(s, "Response") },
			PatternType: PatternTypeStructLiterals,
			Description: "settings gRPC converters",
		},
		{
			SourcePkg:   "internal/domain/notifications",
			TargetPkgs:  []string{"internal/services/notifications/grpc/converters"},
			TypeFilter:  func(s string) bool { return strings.Contains(s, "Input") || strings.Contains(s, "Response") },
			PatternType: PatternTypeStructLiterals,
			Description: "notifications gRPC converters",
		},
		{
			SourcePkg:   "internal/domain/audit",
			TargetPkgs:  []string{"internal/services/audit/grpc/converters"},
			TypeFilter:  func(s string) bool { return strings.Contains(s, "Input") || strings.Contains(s, "Response") },
			PatternType: PatternTypeStructLiterals,
			Description: "audit gRPC converters",
		},
	}

	totalCount := 0
	for _, config := range configs {
		for _, targetPkg := range config.TargetPkgs {
			count, err := comparePackages(config, targetPkg)
			totalCount += count
			if err != nil {
				errors = multierror.Append(errors, err)
			}
		}
	}

	log.Printf("Checked %d struct(s) total", totalCount)

	if errors.ErrorOrNil() != nil {
		log.Fatal(errors)
	}
}
