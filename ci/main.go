// CI Pipelines
//
// This module is used to build everything for the dagger-module-helm.

package main

import (
	"context"
	"dagger/ci/internal/dagger"
)

type Ci struct{}

// Build and publish image from Dockerfile using a build context directory
// in a different location than the current working directory
func (m *Ci) Build(
	ctx context.Context,
) (*dagger.Container) {

	// location of source directory
	src := dag.CurrentModule().Source().Directory(".")

	// location of Containerfile
	containerfile := dag.CurrentModule().Source().File("Containerfile.helm")

	// get build context with dockerfile added
	workspace := dag.Container().
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithFile("/src/custom.Dockerfile", containerfile).
		Directory("/src")

	// build using Dockerfile and publish to registry
	//ref, err := dag.Container().
	return dag.Container().
		Build(workspace, dagger.ContainerBuildOpts{
			Dockerfile: "custom.Dockerfile",
		}) // .Publish(ctx, "ttl.sh/hello-dagger")
}
