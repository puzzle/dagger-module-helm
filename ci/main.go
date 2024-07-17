// CI Pipelines
//
// This module is used to build everything for the dagger-module-helm.

package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
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

	// build using Dockerfile
	return dag.Container().
		Build(workspace, dagger.ContainerBuildOpts{
			Dockerfile: "custom.Dockerfile",
		})
}

func (m *Ci) Publish(
	ctx context.Context,
	// URL of the registry
	registry string,
	// name of the repository
	repository string,
	// tag of the image
	// +optional
	tag string,
	// registry user name
	username string,
	// registry user password
	password *dagger.Secret,
) (string, error) {

	container := m.Build(ctx)

	imageTag := "latest"
	if tag != "" {
		imageTag = tag
	}

	reference := fmt.Sprintf("%s/%s:%s", registry, repository, imageTag)

	// publish to registry
	ref, err := container.
		WithRegistryAuth(registry, username, password).
		Publish(ctx, reference)

	if err != nil {
		return "", err
	}

	return ref, nil
}
