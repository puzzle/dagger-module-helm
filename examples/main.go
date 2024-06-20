package main

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

type Examples struct{}

// All executes all tests.
func (m *Examples) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.Version)

	return p.Wait()
}

func (m *Examples) Version(
	// method call context
	ctx context.Context,
) error {
	const expected = "0.1.1"

	// dagger call version --directory ./mychart/
	directory := dag.CurrentModule().Source().Directory("testdata/mychart/")
	version, err := dag.Helm().Version(ctx, directory)

	if err != nil {
		return err
	}

	if version != expected {
		return fmt.Errorf("expected %q, got %q", expected, version)
	}

	return nil
}
