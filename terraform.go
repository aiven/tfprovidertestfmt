package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type cleanupFuncs []func() error

func (fs cleanupFuncs) run() {
	log.Println("[INFO] running cleanup")
	for i := range fs {
		if err := fs[i](); err != nil {
			log.Println("[ERROR] cleanup failed:", err)
		}
	}
}

func setupTerraform(ctx context.Context) (*tfexec.Terraform, cleanupFuncs, error) {
	var (
		err     error
		cleanup = make(cleanupFuncs, 0)
	)
	defer func() {
		if err != nil {
			cleanup.run()
		}
	}()

	installer := &releases.LatestVersion{
		Product: product.Terraform,
	}
	cleanup = append(cleanup, func() error { return installer.Remove(ctx) })

	execPath, err := installer.Install(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to install terraform: %w", err)
	}

	workingDir, err := os.MkdirTemp(os.TempDir(), "tfproviderfmt-")
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create exec dir for terraform: %w", err)
	}
	cleanup = append(cleanup, func() error { return os.RemoveAll(workingDir) })

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create terraform handle: %w", err)
	}
	return tf, cleanup, nil
}
