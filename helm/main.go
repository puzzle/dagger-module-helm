// Provides Helm functionality
//
// The main focus is to publish new Helm Chart versions on registries.
//
// For accomplishing this the https://helm.sh/ tool is used.

package main

import (
	"context"
	"dagger/helm/internal/dagger"
	"fmt"
	"os"
	"strings"
)

const HELM_IMAGE string = "quay.io/puzzle/dagger-module-helm:latest"

type Helm struct{}

type PushOpts struct {
	Registry   		string `yaml:"registry"`
	Repository 		string `yaml:"repository"`
	Oci        		bool   `yaml:"oci"`
	Username   		string `yaml:"username"`
	Password   		*dagger.Secret
	Version    		string `yaml:"version"`
	AppVersion		string `yaml:"appVersion"`
}

func (p PushOpts) getProtocol() string {
	if p.Oci {
		return "oci"
	} else {
		return "https"
	}
}

func (p PushOpts) getRepoFqdn() string {
	return fmt.Sprintf("%s://%s/%s", p.getProtocol(),  p.Registry, p.Repository)
}

func (p PushOpts) getChartFqdn(name string) string {
	return fmt.Sprintf("%s/%s", p.getRepoFqdn(), name)
}

func (p PushOpts) getHelmPkgCmd() []string {
	helmPkgCmd := []string{"helm", "package", "."}
	if p.Version != "" {
		helmPkgCmd = append(helmPkgCmd, "--version", p.Version)
	}
	if p.AppVersion != "" {
		helmPkgCmd = append(helmPkgCmd, "--app-version", p.AppVersion)
	}
	return helmPkgCmd
}

// Get and display the name of the Helm Chart located inside the given directory.
//
// Example usage: dagger call name --directory ./helm/examples/testdata/mychart/
func (h *Helm) Name(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
) (string, error) {
	return h.queryChartWithYq(ctx, directory, ".name")
}

// Get and display the version of the Helm Chart located inside the given directory.
//
// Example usage: dagger call version --directory ./helm/examples/testdata/mychart/
func (h *Helm) Version(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
) (string, error) {
	return h.queryChartWithYq(ctx, directory, ".version")
}

// Get and display the appVersion of the Helm Chart located inside the given directory.
//
// Example usage: dagger call app-version --directory ./helm/examples/testdata/mychart/
func (h *Helm) AppVersion(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
) (string, error) {
	return h.queryChartWithYq(ctx, directory, ".appVersion")
}

// Packages and pushes a Helm chart to a specified OCI-compatible (by default) registry with authentication.
//
// Returns true if the chart was successfully pushed, or false if the chart already exists, with error handling for
// push failures.
//
// Example usage:
//
//	dagger call package-push \
//	  --registry registry.puzzle.ch \
//	  --repository helm \
//	  --username $REGISTRY_HELM_USER \
//	  --password env:REGISTRY_HELM_PASSWORD \
//	  --directory ./examples/testdata/mychart/
//
// Example usage for pushing to a legacy (non-OCI) Helm repository assuming the repo name is 'helm'. If your target URL
// requires a vendor-specific path prefix (for example, JFrog Artifactory usually requires 'artifactory' before the repo
// name) then add it before the repository name. If you want to put the chart in a subpath in
// the repository, then append that to the end of the repository name.
//
//	dagger call package-push \
//		--registry registry.puzzle.ch \
//		--repository vendor-specific-prefix/helm/optional/subpath/in/repository \
//		--username $REGISTRY_HELM_USER \
//		--password env:REGISTRY_HELM_PASSWORD \
//		--directory ./examples/testdata/mychart/ \
//		--use-non-oci-helm-repo=true
func (h *Helm) PackagePush(
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
	// use a non-OCI (legacy) Helm repository
	// +optional
	// +default=false
	useNonOciHelmRepo bool,    // Dev note: We must use default=false due to https://github.com/dagger/dagger/issues/8810
	// set chart version when packaging
	// +optional
	// default=""
	setVersionTo string,
	// set chart appVersion when packaging
	// +optional
	// default=""
	setAppVersionTo string,
) (bool, error) {
	version, err := h.valueOrQueryChart(setVersionTo, ctx, directory, ".version")
	if err != nil {
		return false, err
	}
	opts := PushOpts{
		Registry:   registry,
		Repository: repository,
		Oci:        !useNonOciHelmRepo,
		Username:   username,
		Password:   password,
		Version:    version,
		AppVersion: setAppVersionTo,
	}

	fmt.Fprintf(os.Stdout, "☸️ Helm package and Push")
	c := dag.Container().
		From("registry.puzzle.ch/cicd/alpine-base:latest").
		WithDirectory("/helm", directory).
		WithWorkdir("/helm")

	name, err := h.Name(ctx, directory)
	if err != nil {
		return false, err
	}

	err = h.ociRegistryLoginIfNeeded(ctx, c, &opts)
	if err != nil {
		return false, err
	}

	pkgFile := fmt.Sprintf("%s-%s.tgz", name, version)

	chartExists, err := h.doesChartExistOnRepo(ctx, c, &opts, name, version)
	if err != nil {
		return false, err
	}

	if chartExists {
		return false, nil
	}

	c, err = c.WithExec([]string{"helm", "dependency", "update", "."}).
		WithExec(opts.getHelmPkgCmd()).
		WithExec([]string{"sh", "-c", "ls"}).
		Sync(ctx)

	if err != nil {
		return false, err
	}

	c = c.
		WithEnvVariable("REGISTRY_USERNAME", opts.Username).
		WithSecretVariable("REGISTRY_PASSWORD", opts.Password)

	if useNonOciHelmRepo {
		curlCmd := []string{
			`sh`,
			`-c`,
			`curl --variable %REGISTRY_USERNAME`,
			`--variable %REGISTRY_PASSWORD`,
			`--expand-user "{{REGISTRY_USERNAME}}:{{REGISTRY_PASSWORD}}"`,
			`-T`,
			pkgFile,
			opts.getRepoFqdn() + "/",
		}

		c, err = c.WithExec(curlCmd).Sync(ctx)
	} else {
		registryLoginCmd := []string{
			`sh`,
			`-c`,
			`echo ${REGISTRY_PASSWORD}`, `|`,
			`helm registry login ${REGISTRY_URL}`,
			`--username ${REGISTRY_USERNAME}`,
			`--password-stdin`,
		}
		c, err = c.
			WithEnvVariable("REGISTRY_URL", opts.Registry).
			WithEnvVariable("REGISTRY_USERNAME", opts.Username).
			WithSecretVariable("REGISTRY_PASSWORD", opts.Password).
			WithExec(registryLoginCmd).
			WithExec([]string{"helm", "push", pkgFile, opts.getRepoFqdn()}).
			WithoutSecretVariable("REGISTRY_PASSWORD").
			Sync(ctx)
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// Run Helm unittests with the given directory and files.
//
// Provide the helm chart directory with pointing to it with the `--directory` flag.
// Add the directory location with `"."` as `--args` parameter to tell helm unittest where to find the helm chart with
// the tests.
//
// Example usage: dagger call test --directory ./helm/examples/testdata/mychart/ --args "."
func (h *Helm) Test(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
	// Helm Unittest arguments
	args []string,
) (string, error) {
	c := h.createContainer(directory)
	out, err := c.WithExec([]string{"sh", "-c",
										fmt.Sprintf("%s %s", "helm-unittest", strings.Join(args, " "))}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return out, nil
}

// Run Helm lint with the given directory.
//
// Provide the helm chart directory with pointing to it with the `--directory` flag.
// Use `--args` parameter to pass alternative chart locations or additional options to Helm lint - see
// https://helm.sh/docs/helm/helm_lint/#options
//
// Example usage: dagger call lint --directory ./helm/examples/testdata/mychart/ --args "--quiet"
func (h *Helm) Lint(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
	// Helm lint arguments
	// +optional
	args []string,
) (string, error) {
	var c *dagger.Container

	if h.hasMissingDependencies(ctx, directory) {
		c = h.createContainer(directory).WithMountedDirectory("./charts", h.dependencyUpdate(directory))
	} else {
		c = h.createContainer(directory)
	}

	out, err := c.WithExec([]string{"sh", "-c", fmt.Sprintf("%s %s", "helm lint", strings.Join(args, " "))}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (h *Helm) ociRegistryLoginIfNeeded(
	ctx context.Context,
	c *dagger.Container,
	opts *PushOpts,
) error {
	if !opts.Oci {
		return nil
	}
	registryLoginCmd := []string{
		`sh`,
		`-c`,
		`echo ${REGISTRY_PASSWORD}`, `|`,
		`helm registry login ${REGISTRY_URL}`,
		`--username ${REGISTRY_USERNAME}`,
		`--password-stdin`,
	}
	c, err := c.
		WithEnvVariable("REGISTRY_URL", opts.Registry).
		WithEnvVariable("REGISTRY_USERNAME", opts.Username).
		WithSecretVariable("REGISTRY_PASSWORD", opts.Password).
		WithExec(registryLoginCmd).
		WithoutSecretVariable("REGISTRY_PASSWORD").
		Sync(ctx)
	return err
}

func (h *Helm) doesChartExistOnRepo(
	ctx context.Context,
	c *dagger.Container,
	opts *PushOpts,
	name string,
	version string,
) (bool, error) {
	if opts.Oci {
		// We assume we are already logged into an OCI registry
		c, err := c.WithExec([]string{"sh", "-c", fmt.Sprintf("helm show chart %s --version %s; echo -n $? > /ec", opts.getChartFqdn(name), version)}).Sync(ctx)
		if err != nil {
			return false, err
		}

		exc, err := c.File("/ec").Contents(ctx)
		if err != nil {
			return false, err
		}

		if exc == "0" {
			//Chart exists
			return true, nil
		} 

		return false, nil
	}
	// else non-OCI
	pkgFile := fmt.Sprintf("%s-%s.tgz", name, version)
	// Do a GET of the chart but with response headers only so we do not download the chart
	curlCmd := []string{
		`sh`,
		`-c`,
		`curl --variable %REGISTRY_USERNAME`,
			 `--variable %REGISTRY_PASSWORD`,
			 `--expand-user "{{REGISTRY_USERNAME}}:{{REGISTRY_PASSWORD}}"`,
			 opts.getChartFqdn(pkgFile),
			 `--output /dev/null`,
			 `--silent -Iw '%{http_code}'`,
	}

	httpCode, err := c.
		WithEnvVariable("REGISTRY_USERNAME", opts.Username).
		WithSecretVariable("REGISTRY_PASSWORD", opts.Password).
		WithExec(curlCmd).
		WithoutSecretVariable("REGISTRY_PASSWORD").
		Stdout(ctx)

	if err != nil {
		return false, err
	}

	httpCode = strings.TrimSpace(httpCode)

	if httpCode == "200" {
		// Chart exists
		return true, nil
	}

	if httpCode == "404" {
		return false, nil
	}

	return false, fmt.Errorf("Server returned error code %s checking for chart existence on server.", httpCode)
}

func (h *Helm) queryChartWithYq(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
	yqQuery string,
) (string, error) {
	c := h.createContainer(directory)
	version, err := c.WithExec([]string{"sh", "-c", fmt.Sprintf(`helm show chart . | yq eval '%s' -`, yqQuery)}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(version), nil
}

func (h *Helm) hasMissingDependencies(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
) bool {
	_, err := h.createContainer(directory).WithExec([]string{"sh", "-c", "helm dependency list | grep missing"}).Stdout(ctx)
	return err == nil
}

func (h *Helm) dependencyUpdate(
	// directory that contains the Helm Chart
	directory *dagger.Directory,
) *dagger.Directory {
	c := h.createContainer(directory)
	return c.WithExec([]string{"sh", "-c", "helm dependency update"}).Directory("charts")
}

func (h *Helm) createContainer(
	// directory that contains the Helm Chart
	directory *dagger.Directory,
) *dagger.Container {
	return dag.Container().
		From(HELM_IMAGE).
		WithDirectory("/helm", directory, dagger.ContainerWithDirectoryOpts{Owner: "1001"}).
		WithWorkdir("/helm").
		WithoutEntrypoint()
}

// coalesce returns the first non-empty string from the provided arguments.
func (h *Helm) valueOrQueryChart(
	theValue string, 
	ctx context.Context, 
	directory *dagger.Directory, 
	yqQuery string,
) (string, error) {
	if theValue != "" {
		return strings.TrimSpace(theValue), nil
	}
	return h.queryChartWithYq(ctx, directory, yqQuery)
}
