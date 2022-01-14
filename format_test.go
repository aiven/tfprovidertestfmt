package main

import (
	"context"
	"go/token"
	"reflect"
	"testing"
)

func TestFormatEmbeddedTerraformManifest(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		expect  []byte
	}{
		{
			name: "simple",
			content: []byte(wrapManifestInVar(`
resource "foo" {
  attr1 = foo
     attr2  = bar
      }
`)),
			expect: []byte(wrapManifestInVar(`
resource "foo" {
  attr1 = foo
  attr2 = bar
}`)),
		},
	}

	var ctx = context.Background()

	fs := token.NewFileSet()

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatEmbeddedTerraformManifests(ctx, fs, tt.content)
			if err != nil {
				t.Fatal("unexpected error formatting file:", err)
			}
			if !reflect.DeepEqual(tt.expect, got) {
				t.Fatalf("expected %s, got %s", tt.expect, got)
			}
		})
	}
}

func wrapManifestInVar(tfManifest string) string {
	return `package pkg

var mfst = ` + "`" + tfManifest + "`\n"
}
