// Go examples for the Helm module.
//
// This module defines the examples for the Daggerverse.	

package main

import (
	"context"
	"dagger/examples/internal/dagger"
)

type Examples struct{}


// Example on how to call the PackagePush method.
// Packages and pushes a Helm chart to a specified OCI-compatible registry with authentication.
//
// Return: true if the chart was successfully pushed, or false if the chart already exists, with error handling for push failures.
func (h *Examples) HelmPackagepush(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
	// URL of the registry
	registry string,
	// name of the repository
	repository string,
	// registry login username
	username string,
	// registry login password
	password *dagger.Secret,
) (bool, error) {
	return dag.
			Helm().
			PackagePush(ctx, directory, registry, repository, username, password)
}

// Example on how to call the Test method.
//
// Run the unit tests for the Helm Chart located inside the directory referenced by the directory parameter.
// Add the directory location with `"."` as `--args` parameter to tell helm unittest where to find the tests inside the passed directory.
//
// Return: The Helm unit test output as string.
func (h *Examples) HelmTest(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart, e.g. "./helm/examples/testdata/mychart/"
	directory *dagger.Directory,
	// Helm Unittest arguments, e.g. "." to reference the Helm Chart root directory inside the passed directory.
	args []string,
) (string, error) {
	return dag.
			Helm().
			Test(ctx, directory, args)
}

// Example on how to call the Version method.
// 
// Get and display the version of the Helm Chart located inside the directory referenced by the directory parameter.
//
// Return: The Helm Chart version as string.
func (m *Examples) HelmVersion(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart, e.g. "./helm/examples/testdata/mychart/"
	chart *dagger.Directory,
) (string, error) {
	return dag.
			Helm().
			Version(ctx, chart)
}
