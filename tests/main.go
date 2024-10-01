
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

	return p.Wait()
}

func (m *Go) HelmVersion(
	// method call context
	ctx context.Context,
) error {
	const expected = "0.1.1"

	// dagger call version --directory ./examples/testdata/mychart/
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
