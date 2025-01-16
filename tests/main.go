package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"

	"github.com/sourcegraph/conc/pool"
)

type Go struct{}

// All executes all tests.
func (m *Go) All(ctx context.Context) error {
	p := pool.New().WithErrors().WithContext(ctx)

	p.Go(m.HelmVersion)
	p.Go(m.HelmTest)
	p.Go(m.HelmLint)
	p.Go(m.HelmLintWithArg)
	p.Go(m.HelmLintWithArgs)
	p.Go(m.HelmLintWithMissingDependencies)

	return p.Wait()
}

func (m *Go) HelmVersion(
	// method call context
	ctx context.Context,
) error {
	const expected = "0.1.1"
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	version, err := dag.Helm().Version(ctx, directory)

	if err != nil {
		return err
	}

	if version != expected {
		return fmt.Errorf("expected %q, got %q", expected, version)
	}

	return nil
}

func (m *Go) HelmPackagepush(
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
	// directory that contains the Helm Chart
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	//service := m.RegistryService(ctx)
	service := m.RegistryContainer(ctx).AsService().WithHostname("localregistry")
	service.Start(ctx)
	_, err := dag.Helm().PackagePush(ctx, directory, registry, repository, username, password)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) RegistryContainer(
	ctx context.Context,
) *dagger.Container {
	return dag.Container().
		From("registry:latest").
		WithExposedPort(5000)
}

func (m *Go) RegistryService(
	ctx context.Context,
) *dagger.Service {
	return dag.Container().
		From("registry:latest").
		WithExposedPort(5000).
		AsService().
		WithHostname("localregistry")
	// service, err := service.Start(ctx)
	// if err != nil {
	// 	return nil
	// }
	// return service
}

func (m *Go) Build(
	_ context.Context,
	dir *dagger.Directory,
) *dagger.Container {
	return dag.Container().
		From("golang:1.23").
		WithDirectory("/src", dir).
		WithWorkdir("/src").
		//WithExec([]string{"apt", "update"}).
		//WithExec([]string{"apt", "install", "docker.io", "-y"}).
		//WithExec([]string{"curl -L https://dl.dagger.io/dagger/install.sh", "| BIN_DIR=/src sh"}).
		WithEnvVariable("BIN_DIR", "/src/bin").
		WithExec([]string{"mkdir", "bin/"}).
		WithExec([]string{"curl", "-fsSL", "https://dl.dagger.io/dagger/install.sh", "-o", "install.sh"}).
		WithExec([]string{"sh", "install.sh"}).
		WithExec([]string{"./bin/dagger", "version"}).
		WithExec([]string{"./bin/dagger", "install", "helm"})
}

func (m *Go) HelmTest(
	// method call context
	ctx context.Context,
) error {
	args := []string{"."}
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	_, err := dag.Helm().Test(ctx, directory, args)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmLint(
	// method call context
	ctx context.Context,
) error {
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	_, err := dag.Helm().Lint(ctx, directory)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmLintWithArg(
	// method call context
	ctx context.Context,
) error {
	args := dagger.HelmLintOpts{Args: []string{"--quiet"}}
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	_, err := dag.Helm().Lint(ctx, directory, args)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmLintWithArgs(
	// method call context
	ctx context.Context,
) error {
	args := dagger.HelmLintOpts{Args: []string{"--strict", "--quiet"}}
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	_, err := dag.Helm().Lint(ctx, directory, args)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmLintWithMissingDependencies(
	// method call context
	ctx context.Context,
) error {
	directory := dag.CurrentModule().Source().Directory("./testdata/mydependentchart/")
	_, err := dag.Helm().Lint(ctx, directory)

	if err != nil {
		return err
	}

	return nil
}
