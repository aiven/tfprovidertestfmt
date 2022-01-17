package main

import (
	"go/token"
	"io/ioutil"
	"path"
	"reflect"
	"testing"
)

func TestFormatEmbeddedTerraformManifest(t *testing.T) {
	tests := []string{
		"variable",
		"func",
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt, func(t *testing.T) {
			input, err := ioutil.ReadFile(path.Join("test", tt+".input.go"))
			if err != nil {
				t.Fatal("unable to read input file:", err)
			}
			expect, err := ioutil.ReadFile(path.Join("test", tt+".expect.go"))
			if err != nil {
				t.Fatal("unable to read expectation file:", err)
			}

			formatted, err := formatEmbeddedTerraformManifests(token.NewFileSet(), input)
			if err != nil {
				t.Fatal("unexpected error formatting file:", err)
			}
			if !reflect.DeepEqual(expect, formatted) {
				t.Fatalf("expected %s, got %s", expect, formatted)
			}
		})
	}
}
