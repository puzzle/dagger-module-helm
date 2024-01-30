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

// Read version from Helm Chart
func (h *Helm) Version(ctx context.Context, directory *Directory) (string, error) {
	c := dag.Container().From("registry.puzzle.ch/cicd/alpine-base:latest").WithDirectory("/helm", directory).WithWorkdir("/helm")
	version, err := c.WithExec([]string{"sh", "-c", "helm show chart . | yq eval '.version' -"}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(version), nil
}

// Package and push an Helm Chart into a registry
// return true && nil, chart was pushed successfully
// return false && nil, chart was not pushed. The specified version already exists
// return true/false && error, an error occurred
func (h *Helm) PackagePush(ctx context.Context, directory *Directory, registry string, repository string, username string, password string) (bool, error) {

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

// Run helm unit test in a given directory
func (h *Helm) Test(
	// method call context
	ctx context.Context,

	// local directory that contains the Helm Chart
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
