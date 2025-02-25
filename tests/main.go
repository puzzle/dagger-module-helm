package main

import (
	"context"
	"dagger/go/internal/dagger"
	"fmt"
	"time"

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
	p.Go(m.HelmPackagePush)
	p.Go(m.HelmPackagePushNonOci)
	p.Go(m.HelmPackagePushWithExistingChart)

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

// requires valid credentials, called from Github actions
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
	randomString := fmt.Sprintf("%d", time.Now().UnixNano())
	// directory that contains the Helm Chart
	directory := m.chartWithVersionSuffix(dag.CurrentModule().Source().Directory("./testdata/mychart/"), randomString)
	_, err := dag.Helm().PackagePush(ctx, directory, registry, repository, username, password)

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmPackagePush(
	// method call context
	ctx context.Context,
) error {
	// directory that contains the Helm Chart
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	_, err := dag.Helm().PackagePush(ctx, directory, "ttl.sh", "helm", "username", dag.SetSecret("password", "secret"))

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmPackagePushNonOci(
	// method call context
	ctx context.Context,
) error {
	// directory that contains the Helm Chart
	directory := dag.CurrentModule().Source().Directory("./testdata/mychart/")
	_, err := dag.Helm().PackagePush(ctx, directory, "ttl.sh", "helm", "username", dag.SetSecret("password", "secret"), dagger.HelmPackagePushOpts{UseNonOciHelmRepo: true})

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmPackagePushWithExistingChart(
	// method call context
	ctx context.Context,
) error {
	randomString := fmt.Sprintf("%d", time.Now().UnixNano())
	// directory that contains the Helm Chart
	directory := m.chartWithVersionSuffix(dag.CurrentModule().Source().Directory("./testdata/mychart/"), randomString)

	returnValue, err := dag.Helm().PackagePush(ctx, directory, "ttl.sh", "helm", "username", dag.SetSecret("password", "secret"))
	if err != nil {
		return err
	}
	if !returnValue {
		return fmt.Errorf("should return true because chart does not exists")
	}

	returnValue, err = dag.Helm().PackagePush(ctx, directory, "ttl.sh", "helm", "username"+randomString, dag.SetSecret("password", "secret"))
	if err != nil {
		return err
	}
	if returnValue {
		return fmt.Errorf("should return false because chart already exists")
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

func (m *Go) chartWithVersionSuffix(
	// directory that contains the Helm Chart
	directory *dagger.Directory,
	randomString string,
) *dagger.Directory {
	// set name and version to arbitrary values
	directory = directory.
		WithoutFile("Chart.yaml").
		WithNewFile("Chart.yaml",
			fmt.Sprintf(`
apiVersion: v2
name: dagger-module-helm-test
description: A Helm chart
type: application
version: 0.1.%s
`, randomString))

	return directory
}
