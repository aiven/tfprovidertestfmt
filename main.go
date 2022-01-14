package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/token"
	"io/ioutil"
	"log"
	"os"
)

var (
	lintFlag        = flag.Bool("lint", false, "check if all files are formatted properly")
	inplaceFlag     = flag.Bool("inplace", false, "format all files in place")
	indentationFlag = flag.Bool("indent", false, "indent the embedded manifests to the appropriate level")
)

func main() {
	flag.Parse()

	if err := run(context.Background()); err != nil {
		log.Fatal("[ERROR]: ", err)
	}
}

func run(ctx context.Context) error {
	if *lintFlag && *inplaceFlag || !*lintFlag && !*inplaceFlag {
		return fmt.Errorf("must set either 'lint' or 'inplace'")
	}

	log.Println("[INFO] looking at files")
	for _, path := range flag.Args() {
		log.Println("[INFO] looking at", path)
		if err := handleFile(ctx, path); err != nil {
			return fmt.Errorf("unable to handle file at path: %s: %w", path, err)
		}
	}
	log.Println("[INFO] looked at all files")
	return nil
}

func handleFile(ctx context.Context, path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("unable to open file at path '%s': %w", path, err)
	}
	defer func() { _ = file.Close() }()

	fset := token.NewFileSet()

	oldContent, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("unable to read file at path '%s': %w", path, err)
	}

	newContent, err := formatEmbeddedTerraformManifests(ctx, fset, oldContent)
	if err != nil {
		return fmt.Errorf("unable to format file at path '%s': %w", path, err)
	}

	if *lintFlag && !bytes.Equal(oldContent, newContent) {
		return fmt.Errorf("file at path '%s' is not formatted", path)
	}

	if err = file.Truncate(0); err != nil {
		return fmt.Errorf("unable to truncate file at path '%s': %w", path, err)
	}
	if _, err = file.Seek(0, 0); err != nil {
		return fmt.Errorf("unable to seek beginning of file at path '%s': %w", path, err)
	}
	if _, err = file.Write(newContent); err != nil {
		return fmt.Errorf("unable to write formatted content to path '%s': %w", path, err)
	}
	return nil
}
