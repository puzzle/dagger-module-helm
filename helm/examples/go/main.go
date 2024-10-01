// Go examples for the Helm module.
//
// This module defines the examples for the Daggerverse.	

package main

import (
	"context"
	"dagger/go/internal/dagger"
)

type Go struct{}

func (h *Go) HelmPackagepush(
	// method call context
	ctx context.Context,
	// URL of the registry
	registry string,
	// name of the repository
	repository string,
	// registry login username
	username string,
	// registry login password
	password *dagger.Secret,
) error {
	//	dagger call package-push \
	//	  --registry registry.puzzle.ch \
	//	  --repository helm \
	//	  --username $REGISTRY_HELM_USER \
	//	  --password env:REGISTRY_HELM_PASSWORD \
	//	  --directory ./examples/testdata/mychart/

	// directory that contains the Helm Chart
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	_, err := dag.Helm().PackagePush(ctx, directory, registry, repository, username, password)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmTest(
	// method call context
	ctx context.Context,
) error {
	args := []string{"."}

	// dagger call test --directory ./examples/testdata/mychart/ --args "."
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	_, err := dag.Helm().Test(ctx, directory, args)

	if err != nil {
		return err
	}

	return nil
}

// Example on how to call the Version method.
// 
// Get and display the version of the Helm Chart located inside the directory referenced by the chart parameter.
func (m *Go) HelmVersion(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart, e.g. "./tests/testdata/mychart/"
	chart *dagger.Directory,
) (string, error) {
	return dag.
			Helm().
			Version(ctx, chart)
}
