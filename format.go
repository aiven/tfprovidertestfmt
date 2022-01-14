package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"unicode"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func formatEmbeddedTerraformManifests(fset *token.FileSet, content []byte) ([]byte, error) {
	parsed, err := parser.ParseFile(fset, "src.go", bytes.NewReader(content), parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("unable to parse content: %w", err)
	}

	var stack []ast.Node
	ast.Inspect(parsed, func(n ast.Node) bool {
		defer func() {
			if n == nil {
				// Done with node's children. Pop.
				stack = stack[:len(stack)-1]
			} else {
				// Push the current node for children.
				stack = append(stack, n)
			}
		}()

		if n == nil || len(stack) == 0 {
			return true
		}

		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if len(lit.Value) == 0 || lit.Value[0] != '`' {
			return true
		}

		val := strings.Trim(lit.Value, "`")
		val = strings.ReplaceAll(val, "\t", "")

		if !isTerraformManifest(val) {
			return true
		} else {
			val = formatTerraformManifest(val)
		}
		indentation := getExpectedIndentation(stack, fset)

		lit.Value = "`" + strings.TrimRightFunc(strings.ReplaceAll(val, "\n", "\n"+strings.Repeat("\t", indentation)), unicode.IsSpace) + "`"
		return true

	})

	buf := new(bytes.Buffer)

	if err = format.Node(buf, fset, parsed); err != nil {
		return nil, fmt.Errorf("unable to write formatted output into buffer: %w", err)
	}
	return buf.Bytes(), nil
}

func isTerraformManifest(s string) bool {
	_, diag := hclparse.NewParser().ParseHCL([]byte(s), "cand.tf")
	return !diag.HasErrors()
}

func formatTerraformManifest(s string) string {
	return string(hclwrite.Format([]byte(s)))
}

func getExpectedIndentation(stack []ast.Node, fset *token.FileSet) int {
	if ret := lastReturnStatement(stack); ret != nil {
		return fset.Position(ret.Pos()).Column
	}
	if ass := lastAssignmentStatement(stack); ass != nil {
		return fset.Position(ass.Pos()).Column
	}
	if ass := lastDeclarationStatement(stack); ass != nil {
		return fset.Position(ass.Pos()).Column
	}
	if ass := lastValueSpec(stack); ass != nil {
		return fset.Position(ass.Pos()).Column - 4
	}
	return 0
}

func lastReturnStatement(stack []ast.Node) ast.Node {
	for i := len(stack) - 1; i != 0; i-- {
		switch stack[i].(type) {
		case *ast.ReturnStmt:
			return stack[i]
		default:
			continue
		}
	}
	return nil
}

func lastAssignmentStatement(stack []ast.Node) ast.Node {
	for i := len(stack) - 1; i != 0; i-- {
		switch stack[i].(type) {
		case *ast.AssignStmt:
			return stack[i]
		default:
			continue
		}
	}
	return nil
}

func lastDeclarationStatement(stack []ast.Node) ast.Node {
	for i := len(stack) - 1; i != 0; i-- {
		switch stack[i].(type) {
		case *ast.DeclStmt:
			return stack[i]
		default:
			continue
		}
	}
	return nil
}

func lastValueSpec(stack []ast.Node) ast.Node {
	for i := len(stack) - 1; i != 0; i-- {
		switch stack[i].(type) {
		case *ast.ValueSpec:
			return stack[i]
		default:
			continue
		}
	}
	return nil
}
