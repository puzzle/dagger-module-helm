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
    "hash/crc32"
	"bufio"
)

const HELM_IMAGE string = "quay.io/puzzle/dagger-module-helm:latest"

type Helm struct{}

type PushOpts struct {
	Registry   		string `yaml:"registry"`
	Repository 		string `yaml:"repository"`
	Oci        		bool   `yaml:"oci"`
	NonOCISubpath	string `yaml:"nonOciSubpath"`
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
	if !p.Oci && p.NonOCISubpath != "" {
		return fmt.Sprintf("%s/%s/%s", p.getRepoFqdn(), p.NonOCISubpath, name)
	}
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
// If the chart you want to push has dependencies, then we will assume you will be using the same set of credentials
// given with `--username` and `--password` for each and every dependency in your parent chart and we will automatically
// log into the registry for each dependency if the registry differs from the `--registry` argument.
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
// name) then add it before the repository name. If you want to push the chart into a subpath in the repository, then use
// non-oci-repo-subpath. If the non-oci-repo-subpath option is used, specifying use-non-oci-helm-repo as true is optional.
//
// Example usage without a subpath and with a subpath of 'optional/subpath/in/repository':
//
//	dagger call package-push \
//		--registry registry.puzzle.ch \
//		--repository vendor-specific-prefix/helm \
//		--username $REGISTRY_HELM_USER \
//		--password env:REGISTRY_HELM_PASSWORD \
//		--directory ./examples/testdata/mychart/ \
//		--use-non-oci-helm-repo=true
//
//	dagger call package-push \
//		--registry registry.puzzle.ch \
//		--repository vendor-specific-prefix/helm \
//		--username $REGISTRY_HELM_USER \
//		--password env:REGISTRY_HELM_PASSWORD \
//		--directory ./examples/testdata/mychart/ \
//		--non-oci-repo-subpath optional/subpath/in/repository
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
	// when using a non-OCI Helm repository, optionally specify a subpath location in the repository
	// +optional
	// +default=""
	nonOciRepoSubpath string,
	// set chart version when packaging
	// +optional
	// default=""
	setVersionTo string,
	// set chart appVersion when packaging
	// +optional
	// default=""
	setAppVersionTo string,
) (bool, error) {
	if nonOciRepoSubpath != "" {
		useNonOciHelmRepo = true
		if nonOciRepoSubpath[0] == '/' {
			nonOciRepoSubpath = nonOciRepoSubpath[1:]
		}
	}
	opts := PushOpts{
		Registry:		registry,
		Repository:		repository,
		Oci:			!useNonOciHelmRepo,
		NonOCISubpath:	nonOciRepoSubpath,
		Username:		username,
		Password:		password,
		Version:		setVersionTo,
		AppVersion:		setAppVersionTo,
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

	version := opts.Version
	if version == "" {
		version, err = h.Version(ctx, directory)
		if err != nil {
			return false, err
		}
	}

	pkgFile := fmt.Sprintf("%s-%s.tgz", name, version)

	err = h.registryLogin(ctx, opts.Registry, opts.Username, opts.Password, opts.Oci, c)
	if err != nil {
		return false, err
	}

	chartExists, err := h.doesChartExistOnRepo(ctx, c, &opts, name, version)
	if err != nil {
		return false, err
	}

	if chartExists {
		return false, nil
	}

	missingDependencies, err := h.hasMissingDependencies(ctx, c)
	if err != nil {
		return false, err
	}
	if missingDependencies {
		c, err = h.setupContainerForDependentCharts(ctx, username, password, useNonOciHelmRepo, c, opts.getRepoFqdn())
		if err != nil {
			return false, err
		}

		c, err = c.WithExec([]string{"helm", "dependency", "update", "."}).Sync(ctx)

		if err != nil {
			return false, err
		}
	}

	c, err = c.WithExec(opts.getHelmPkgCmd()).Sync(ctx)
	if err != nil {
		return false, err
	}

	if useNonOciHelmRepo {
		curlCmd := []string{
			`curl --variable %REGISTRY_USERNAME`,
			`--variable %REGISTRY_PASSWORD`,
			`--expand-user "{{REGISTRY_USERNAME}}:{{REGISTRY_PASSWORD}}"`,
			`-T`,
			pkgFile,
			opts.getRepoFqdn() + "/",
		}

		c, err = c.
			WithEnvVariable("REGISTRY_USERNAME", opts.Username).
			WithSecretVariable("REGISTRY_PASSWORD", opts.Password).
			WithExec([]string{"sh", "-c", strings.Join(curlCmd, " ")}).
			WithoutSecretVariable("REGISTRY_PASSWORD").
			Sync(ctx)
	} else {
		c, err = c.
			WithExec([]string{"helm", "push", pkgFile, opts.getRepoFqdn()}).
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
// If you need to be able to pull dependent charts but you need to supply credentials for them, then 
// you can optionally supply the `--username` and `--password` parameters. In this case, this function 
// will assume you will be using the same set of credentials for each and every dependency in your parent 
// chart and will automatically log into the registry for each dependency. If you are using non-OCI Helm repositories,
// you can also specify the `--use-non-oci-helm-repo` parameter to use the legacy Helm repository format.
//
// Example usage without supplying credentials:
//
// dagger call lint --directory ./helm/examples/testdata/mychart/ --args "--quiet"
func (h *Helm) Lint(
	// method call context
	ctx context.Context,
	// directory that contains the Helm Chart
	directory *dagger.Directory,
	// Helm lint arguments
	// +optional
	args []string,
	// supplemental credentials for dependent charts - username
	// +optional
	username string,
	// supplemental credentials for dependent charts - password
	// +optional
	password *dagger.Secret,
	// use a non-OCI (legacy) Helm repository when pulling dependent charts
	// +optional
	// +default=false
	useNonOciHelmRepo bool,    // Dev note: We are forced to use default=false due to https://github.com/dagger/dagger/issues/8810
) (string, error) {
	var c *dagger.Container
	var err error

	c = h.createContainer(directory)
	if username != "" { // setup credentials for dependent charts
		c, err = h.setupContainerForDependentCharts(ctx, username, password, useNonOciHelmRepo, c, "")
		if err != nil {
			return "", err
		}
	}

	missingDependencies, err := h.hasMissingDependencies(ctx, c)
	if err != nil {
		return "", err
	}

	if missingDependencies {
		c = c.WithMountedDirectory("./charts", h.dependencyUpdate(c))
	}

	out, err := c.WithExec([]string{"sh", "-c", fmt.Sprintf("%s %s", "helm lint", strings.Join(args, " "))}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (h *Helm) registryLogin(
	ctx context.Context,
	registry string,
	username string,
	password *dagger.Secret,
	useOciHelmRepo bool,
	c *dagger.Container,
) error {
	c = c.
		WithEnvVariable("REGISTRY_USERNAME", username).
		WithEnvVariable("REGISTRY_URL", registry).
		WithSecretVariable("REGISTRY_PASSWORD", password)

	var cmd []string
	if !useOciHelmRepo {
		// The hash of registry is used as a repository name when setting up non-OCI.
		// The repo name can be anything since we only login to resolve dependencies.
		hash := crc32.ChecksumIEEE([]byte(registry))
		repoHashKey := fmt.Sprintf("%x", hash)
		c = c.WithEnvVariable("REPO_NAME", repoHashKey)
		cmd = []string{
			"sh", "-c",
			`(echo ${REGISTRY_PASSWORD} | helm repo add ${REPO_NAME} ${REGISTRY_URL} --username ${REGISTRY_USERNAME} --password-stdin) ; helm repo update ${REPO_NAME}`,
		}
	} else {
		cmd = []string{
			"sh", "-c", 
			`echo ${REGISTRY_PASSWORD} | helm registry login ${REGISTRY_URL} --username ${REGISTRY_USERNAME} --password-stdin`,
		}
	}

	c, err := c.
		WithExec(cmd).
		WithoutSecretVariable("REGISTRY_PASSWORD").
		Sync(ctx)
	if err != nil {
		return err
	}

	return err
}

func (h *Helm) setupContainerForDependentCharts(
	ctx context.Context,
	username string,
	password *dagger.Secret,
	useNonOciHelmRepo bool,
	c *dagger.Container,
	alreadyLoggedInto string,
) (*dagger.Container, error) {
	c = c.
		WithEnvVariable("REGISTRY_USERNAME", username).
		WithSecretVariable("REGISTRY_PASSWORD", password)

	valuesString, err := c.WithExec([]string{"sh", "-c", `helm show chart . | yq '[.dependencies[].repository] | unique | sort | .[]'`}).Stdout(ctx)
	if err != nil {
		return c, err
	}

	repoURLs := h.getRepoURLs(valuesString)
	for _, repoURL := range repoURLs {
		if repositoriesAreNotEquivalent(alreadyLoggedInto, repoURL) { // avoid logging into the same repo twice
			h.registryLogin(ctx, repoURL, username, password, !useNonOciHelmRepo, c)
			if err != nil {
				return c, err
			}
		}
	}

	return c, nil
}

func repositoriesAreNotEquivalent(
	repositoryOne string,
	repositoryTwo string,
) bool {
	if repositoryOne == "" {
		return true
	}
	normalize := func(input string) string {
		result := strings.TrimSpace(input)
		result = result[:strings.Index(result, "://")]
		result = strings.ToLower(result)
		return result
	}
	return !strings.Contains(normalize(repositoryOne), normalize(repositoryTwo))
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
		WithExec([]string{"sh", "-c", strings.Join(curlCmd, " ")}).
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
	result, err := c.WithExec([]string{"sh", "-c", fmt.Sprintf(`helm show chart . | yq eval '%s' -`, yqQuery)}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

func (h *Helm) hasMissingDependencies(
	// method call context
	ctx context.Context,
	// container to run the command in
	c *dagger.Container,
) (bool, error) {
	output, err := c.WithExec([]string{"sh", "-c", "helm dependency list"}).Stdout(ctx)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "missing") {
		return true, nil
	}
	return false, nil
}

func (h *Helm) dependencyUpdate(
	// directory that contains the Helm Chart
	c *dagger.Container,
) *dagger.Directory {
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

// getRepoURLs returns a list of the repository URLs from the values string.
func (h *Helm) getRepoURLs(
	valuesString string,
) []string {
	resultList := []string{}

    valScanner := bufio.NewScanner(strings.NewReader(valuesString))
    for valScanner.Scan() {
		resultList = append(resultList, valScanner.Text())
    }
    return resultList
}
