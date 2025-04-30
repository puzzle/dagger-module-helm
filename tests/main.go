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

	p.Go(m.HelmName)
	p.Go(m.HelmVersion)
	p.Go(m.HelmAppVersion)
	p.Go(m.HelmTest)
	p.Go(m.HelmLint)
	p.Go(m.HelmLintWithArg)
	p.Go(m.HelmLintWithArgs)
	p.Go(m.HelmLintWithMissingDependencies)
	p.Go(m.HelmPackagePush)

	// We have to simulate a non-OCI helm repository for this test to work...
	//p.Go(m.HelmPackagePushNonOci)

	p.Go(m.HelmPackagePushWithExistingChart)
	p.Go(m.HelmPackagePushWithVersion)

	return p.Wait()
}

func (m *Go) HelmName(
	// method call context
	ctx context.Context,
) error {
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
	expected := Mychart.GetName()

	version, err := dag.Helm().Name(ctx, directory)
	if err != nil {
		return err
	}

	if version != expected {
		return fmt.Errorf("expected %q, got %q", expected, version)
	}

	return nil
}

func (m *Go) HelmVersion(
	// method call context
	ctx context.Context,
) error {
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
	expected := Mychart.GetVersion()

	version, err := dag.Helm().Version(ctx, directory)
	if err != nil {
		return err
	}

	if version != expected {
		return fmt.Errorf("expected %q, got %q", expected, version)
	}

	return nil
}

func (m *Go) HelmAppVersion(
	// method call context
	ctx context.Context,
) error {
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
	version, err := dag.Helm().AppVersion(ctx, directory)
	expected := Mychart.GetAppVersion()

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
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory

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
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
	_, err := dag.
				Helm().
				PackagePush(ctx, directory, "ttl.sh", "helm", "username", dag.SetSecret("password", "secret"))

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmPackagePushNonOci(
	// method call context
	ctx context.Context,
) error {
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
	_, err := dag.
				Helm().
				PackagePush(ctx, directory, "ttl.sh", "helm", "username",
							dag.SetSecret("password", "secret"),
							dagger.HelmPackagePushOpts{UseNonOciHelmRepo: true})

	if err != nil {
		return err
	}

	return nil
}

func (m *Go) HelmPackagePushWithVersion(
	// method call context
	ctx context.Context,
) error {
	Mychart, _ := NewChartSession(ctx)
	const differentVersion = "0.6.7"
	directory := Mychart.DaggerDirectory
	returnValue, err := dag.
				Helm().
				PackagePush(ctx, directory, "ttl.sh", "helm", "username",
							dag.SetSecret("password", "secret"),
							dagger.HelmPackagePushOpts{SetVersionTo: differentVersion})

	if err != nil {
		return err
	}

	if !returnValue {
		return fmt.Errorf("should return true because chart does not exist")
	}

	returnValue, err = dag.
							Helm().
							PackagePush(ctx, directory, "ttl.sh", "helm", "username",
										dag.SetSecret("password", "secret"))
	if err != nil {
		return err
	}
	if !returnValue {
		return fmt.Errorf("should return true because chart does not exist")
	}

	returnValue, err = dag.
				Helm().
				PackagePush(ctx, directory, "ttl.sh", "helm", "username" + Mychart.SessionKey,
							dag.SetSecret("password", "secret"),
							dagger.HelmPackagePushOpts{SetVersionTo: differentVersion})

	if err != nil {
		return err
	}

	if returnValue {
		return fmt.Errorf("should return false because chart already exists")
	}

	return nil
}

func (m *Go) HelmPackagePushWithExistingChart(
	// method call context
	ctx context.Context,
) error {
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory

	returnValue, err := dag.
							Helm().
							PackagePush(ctx, directory, "ttl.sh", "helm", "username",
										dag.SetSecret("password", "secret"))
	if err != nil {
		return err
	}
	if !returnValue {
		return fmt.Errorf("should return true because chart does not exists")
	}

	returnValue, err = dag.
						Helm().
						PackagePush(ctx, directory, "ttl.sh", "helm", "username"+RandomString(8),
									dag.SetSecret("password", "secret"))
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
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
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
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
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
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
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
	Mychart, _ := NewChartSession(ctx)
	directory := Mychart.DaggerDirectory
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
	_, Mydependentchart := NewChartSession(ctx)
	directory := Mydependentchart.DaggerDirectory
	_, err := dag.Helm().Lint(ctx, directory)

	if err != nil {
		return err
	}

	return nil
}

