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

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

func init() {
	flag.Parse()
}

var (
	lintFlag    = flag.Bool("lint", false, "check if all files are formatted properly and exit on violations")
	inplaceFlag = flag.Bool("inplace", false, "format all files in place")
)

func main() {
	ctx := context.Background()

	if *lintFlag && *inplaceFlag || !*lintFlag && !*inplaceFlag {
		log.Fatalf("[ERROR] must set either 'lint' or 'inplace'")
	}

	log.Println("[INFO] installing terraform")

	installer := &releases.LatestVersion{
		Product: product.Terraform,
	}

	execPath, err := installer.Install(ctx)
	if err != nil {
		log.Fatalf("[ERROR] unable to intsall terraform: %s", err)
	}

	tmpdir := os.TempDir()
	defer func() { _ = os.RemoveAll(tmpdir) }()

	tf, err := tfexec.NewTerraform(tmpdir, execPath)
	if err != nil {
		log.Fatalf("[ERROR] unable to create terraform handle: %s", err)
	}
	log.Println("[INFO] installed terraform, looking at files")

	success := true
	for _, path := range flag.Args() {
		log.Println("[INFO] looking at", path)
		if err = handleFile(ctx, path, tf); err != nil {
			success = false
			log.Println("[ERROR]: ", path, ":", err)
		}
	}

	log.Println("[INFO] looked at all files")

	if success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
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
