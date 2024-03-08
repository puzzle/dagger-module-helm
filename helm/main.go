// Provides Helm functionality
//
// The main focus is to publish new Helm Chart versions on registries.
//
// For accomplishing this the https://helm.sh/ tool is used.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type Helm struct {
}

type PushOpts struct {
	Registry   string `yaml:"registry"`
	Repository string `yaml:"repository"`
	Oci        bool   `yaml:"oci"`
	Username   string `yaml:"username"`
	Password   string
}

func (p PushOpts) getChartFqdn(name string) string {
	if p.Oci {
		return fmt.Sprintf("oci://%s/%s/%s", p.Registry, p.Repository, name)
	}
	return fmt.Sprintf("%s/%s/%s", p.Registry, p.Repository, name)
}

func (p PushOpts) getRepoFqdn() string {
	if p.Oci {
		return fmt.Sprintf("oci://%s/%s", p.Registry, p.Repository)
	}
	return fmt.Sprintf("%s/%s", p.Registry, p.Repository)
}

// Get and display the version of the Helm Chart located at the directory given by the `--directory` flag.
//
// Example usage: dagger call version --directory ./mychart/
func (h *Helm) Version(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *Directory,
	) (string, error) {
	c := dag.Container().From("registry.puzzle.ch/cicd/alpine-base:latest").
		WithDirectory("/helm", directory).
		WithWorkdir("/helm")
	version, err := c.WithExec([]string{"sh", "-c", "helm show chart . | yq eval '.version' -"}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(version), nil
}

// Package and push an Helm Chart into a registry
//
// return true && nil, chart was pushed successfully
// return false && nil, chart was not pushed. The specified version already exists
// return true/false && error, an error occurred
//
// Example usage:
// dagger call package-push \
//   --registry registry.puzzle.ch \
//   --repository helm \
//   --username REGISTRY_HELM_USER \
//   --password REGISTRY_HELM_PASSWORD \
//   --directory ./mychart/
func (h *Helm) PackagePush(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *Directory,
	// URL of the registry
	registry string,
	// name of the repository
	repository string,
	// registry login username
	username string,
	// registry login password
	password string,
) (bool, error) {
	opts := PushOpts{
		Registry:   registry,
		Repository: repository,
		Oci:        true,
		Username:   username,
		Password:   password,
	}

	fmt.Fprintf(os.Stdout, "☸️ Helm package and Push")
	c := dag.Container().From("registry.puzzle.ch/cicd/alpine-base:latest").WithDirectory("/helm", directory).WithWorkdir("/helm")
	version, err := c.WithExec([]string{"sh", "-c", "helm show chart . | yq eval '.version' -"}).Stdout(ctx)
	if err != nil {
		return false, err
	}

	version = strings.TrimSpace(version)

	name, err := c.WithExec([]string{"sh", "-c", "helm show chart . | yq eval '.name' -"}).Stdout(ctx)
	if err != nil {
		return false, err
	}

	name = strings.TrimSpace(name)

	c, err = c.WithExec([]string{"helm", "registry", "login", opts.Registry, "-u", opts.Username, "-p", opts.Password}).Sync(ctx)
	if err != nil {
		return false, err
	}

	//TODO: Refactor with return
	c, err = c.WithExec([]string{"sh", "-c", fmt.Sprintf("helm show chart %s --version %s; echo -n $? > /ec", opts.getChartFqdn(name), version)}).Sync(ctx)
	if err != nil {
		return false, err
	}

	exc, err := c.File("/ec").Contents(ctx)
	if err != nil {
		return false, err
	}

	if exc == "0" {
		//Chart exists
		return false, nil
	}

	_, err = c.WithExec([]string{"helm", "dependency", "update", "."}).
		WithExec([]string{"helm", "package", "."}).
		WithExec([]string{"sh", "-c", "ls"}).
		WithExec([]string{"helm", "push", fmt.Sprintf("%s-%s.tgz", name, version), opts.getRepoFqdn()}).
		Sync(ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}


// Run [Helm unittests](https://github.com/helm-unittest/helm-unittest) with the given directory and files.
//
// Provide the helm chart directory with pointing to it with the `--directory` flag.
// Add the directory location with `"."` as `--args` parameter to tell helm unittest where to find the helm chart with the tests.
//
// Example usage: dagger call test --directory ./mychart/ --args "."
func (h *Helm) Test(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *Directory,
	// Helm Unittest arguments
	args []string,
) (string, error) {
	c := dag.Container().From("helmunittest/helm-unittest").WithDirectory("/helm", directory).WithWorkdir("/helm")
	out, err := c.WithExec(args).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return out, nil
}
