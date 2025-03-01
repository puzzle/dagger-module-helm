package main

import (
	"dagger/go/internal/dagger"
	"fmt"
)

const originalVersion = "0.1.1"
const originalAppVersion = "1.16.0"

const mychartName = "dagger-module-helm-test"
const mychartDir = "./testdata/mychart/"
const mychartTemplate = `
apiVersion: v2
name: %s
description: A Helm chart
type: application
version: %s
appVersion: "%s"
`
const mydependentchartName = "dagger-module-helm-test-with-dependency"
const mydependentchartDir = "./testdata/mydependentchart/"
const mydependentchartTemplate = `
apiVersion: v2
name: %s
description: A Helm chart
type: application
version: %s
appVersion: "%s"
dependencies:
  - name: dependency-track
    version: 1.8.1
    repository:  https://puzzle.github.io/dependencytrack-helm/
`

// Chart represents one of our test Helm charts
type Chart struct{
	Name string
	Directory string
	ContentTemplate string
	OriginalContent string
}

// Use the dagger module to get the dagger directory of the chart
func (c *Chart) DaggerDirectory(
) *dagger.Directory {
	return dag.CurrentModule().Source().Directory(c.Directory)
}

// Modify the chart by changing both version and appVersion
func (c *Chart) WithVersionAndAppVersion(
	version string,
	appVersion string,
) *Chart {
	if version == "" {
		version = originalVersion
	}
	if appVersion == "" {
		appVersion = originalAppVersion
	}
	c.DaggerDirectory().
		WithoutFile("Chart.yaml").
		WithNewFile("Chart.yaml",
		fmt.Sprintf(c.ContentTemplate, c.Name, version, appVersion))
	return c
}

// Modify the chart by appending a suffix to the original version
func (c *Chart) WithOriginalVersionSuffix(
	suffix string,
) *Chart {
	versionWithSuffix := fmt.Sprintf("%s-%s", originalVersion, suffix)
	return c.WithVersion(versionWithSuffix)
}

// Modify the chart by changing only the version
func (c *Chart) WithVersion(
	version string,
) *Chart { 
	return c.WithVersionAndAppVersion(version, "") 
}

// Modify the chart by changing only the appVersion
func (c *Chart) WithAppVersion(
	appVersion string,
) *Chart { 
	return c.WithVersionAndAppVersion("", appVersion)
}

// Reset the chart to its original content
func (c *Chart) Reset(
) *Chart {
	return c.WithVersionAndAppVersion("", "")
}

var (
	Mychart = Chart{
		Name: mychartName,
		Directory: mychartDir,
		ContentTemplate: mychartTemplate,
		OriginalContent: fmt.Sprintf(mychartTemplate, mychartName, originalVersion, originalAppVersion),
	}
	Mydependentchart = Chart{
		Name: mydependentchartName,
		Directory: mydependentchartDir,
		ContentTemplate: mydependentchartTemplate,
		OriginalContent: fmt.Sprintf(mydependentchartTemplate, mydependentchartName, originalVersion, originalAppVersion),
	}
)
