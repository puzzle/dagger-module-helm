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
	p.Go(m.Test)

	return p.Wait()
}

func (m *Examples) Version(
	// method call context
	ctx context.Context,
) error {
	const expected = "0.1.1"

	// dagger call version --directory ./examples/testdata/mychart/
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

func (h *Examples) PackagePush(
	// method call context
	ctx context.Context,
	// URL of the registry
	registry string,
	// name of the repository
	repository string,
	// registry login username
	username string,
	// registry login password
	password *Secret,
) error {
	//	dagger call package-push \
	//	  --registry registry.puzzle.ch \
	//	  --repository helm \
	//	  --username $REGISTRY_HELM_USER \
	//	  --password env:REGISTRY_HELM_PASSWORD \
	//	  --directory ./examples/testdata/mychart/

	// directory that contains the Helm Chart
	directory := dag.CurrentModule().Source().Directory("testdata/mychart/")
	_, err := dag.Helm().PackagePush(ctx, directory, registry, repository, username, password)

	if err != nil {
		return err
	}

	return nil
}

func (m *Examples) Test(
	// method call context
	ctx context.Context,
) error {
	args := []string{"."}

	// dagger call test --directory ./examples/testdata/mychart/ --args "."
	directory := dag.CurrentModule().Source().Directory("testdata/mychart/")
	_, err := dag.Helm().Test(ctx, directory, args)

	if err != nil {
		return err
	}

	return nil
}
