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


	p.Go(m.HelmVersion_WithDep)
	p.Go(m.HelmTest_WithDep)
	p.Go(m.HelmLint_WithDep)
	p.Go(m.HelmLintWithArg_WithDep)
	p.Go(m.HelmLintWithArgs_WithDep)
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


// and the same with dependencies

func (m *Go) HelmVersion_WithDep(
	// method call context
	ctx context.Context,
) error {
	const expected = "0.1.1"
	directory := dag.CurrentModule().Source().Directory("./testdata/mychartwithdep/")
	version, err := dag.Helm().Version(ctx, directory)

	if err != nil {
		return err
	}

	if version != expected {
		return fmt.Errorf("expected %q, got %q", expected, version)
	}

	return nil
}

func (h *Go) HelmPackagepush_WithDep(
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
	directory := dag.CurrentModule().Source().Directory("./testdata/mychartwithdep/")
	_, err := dag.Helm().PackagePush(ctx, directory, registry, repository, username, password)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmTest_WithDep(
	// method call context
	ctx context.Context,
) error {
	args := []string{"."}
	directory := dag.CurrentModule().Source().Directory("./testdata/mychartwithdep/")
	_, err := dag.Helm().Test(ctx, directory, args)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmLint_WithDep(
	// method call context
	ctx context.Context,
) error {
	directory := dag.CurrentModule().Source().Directory("./testdata/mychartwithdep/")
	dependencyUpdate := true
	arguments := []string{}
	_, err := dag.Helm().Lint(ctx, directory, arguments)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmLintWithArg_WithDep(
	// method call context
	ctx context.Context,
) error {
	args := dagger.HelmLintOpts{Args: []string{"--quiet"}}
	directory := dag.CurrentModule().Source().Directory("./testdata/mychartwithdep/")
	_, err := dag.Helm().Lint(ctx, directory, args)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmLintWithArgs_WithDep(
	// method call context
	ctx context.Context,
) error {
	args := dagger.HelmLintOpts{Args: []string{"--strict", "--quiet"}}
	directory := dag.CurrentModule().Source().Directory("./testdata/mychartwithdep/")
	_, err := dag.Helm().Lint(ctx, directory, args)

	if err != nil {
		return err
	}

	return nil
}
