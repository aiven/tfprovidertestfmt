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

	"github.com/hashicorp/terraform-exec/tfexec"
)

var (
	lintFlag    = flag.Bool("lint", false, "check if all files are formatted properly")
	inplaceFlag = flag.Bool("inplace", false, "format all files in place")
)

func main() {
	flag.Parse()

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if *lintFlag && *inplaceFlag || !*lintFlag && !*inplaceFlag {
		log.Fatalf("[ERROR] must set either 'lint' or 'inplace'")
	}

	log.Println("[INFO] installing terraform")

	tf, cleanup, err := setupTerraform(ctx)
	if err != nil {
		log.Fatalf("[ERROR] unable to setup terraform: %s", err)
	}
	defer cleanup.run()

	log.Println("[INFO] looking at files")
	for _, path := range flag.Args() {
		log.Println("[INFO] looking at", path)
		if err = handleFile(ctx, path, tf); err != nil {
			return fmt.Errorf("unable to handle file at path: %s: %w", path, err)
		}
	}
	log.Println("[INFO] looked at all files")
	return nil
}

func handleFile(ctx context.Context, path string, tf *tfexec.Terraform) error {
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

	newContent, err := formatEmbeddedTerraformManifests(ctx, fset, tf, oldContent)
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
