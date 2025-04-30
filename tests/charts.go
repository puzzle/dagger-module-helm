package main

import (
	"bytes"
	"dagger/go/internal/dagger"
	"fmt"
	"text/template"
	"crypto/rand"
	"context"
)

const originalVersion = "0.1.1"
const originalAppVersion = "1.16.0"
const randomSuffixLength = 8

const mychartName = "dagger-module-helm-test"
const mychartDir = "./testdata/mychart/"
const mychartTemplateString = `
apiVersion: v2
name: {{ .Name }}
description: A Helm chart
type: application
version: {{ .Version }}
appVersion: "{{ .AppVersion }}"
`
const mydependentchartName = "dagger-module-helm-test-with-dependency"
const mydependentchartDir = "./testdata/mydependentchart/"
const mydependentchartTemplateString = mychartTemplateString + `
dependencies:
  - name: dependency-track
    version: 1.8.1
    repository:  https://puzzle.github.io/dependencytrack-helm/
`

// Chart represents one of our test Helm charts
type Chart struct{
	Name string
	Version string
	AppVersion string
	DaggerDirectory *dagger.Directory
	Directory string
	ContentTemplate *template.Template
	originalContent string
	SessionKey string
}

// Get the name of the chart
func (c Chart) GetName(
) string {
	return c.Name
}

// Get the version of the chart
func (c Chart) GetVersion(
) string {
	return c.Version
}

// Get the appVersion of the chart
func (c Chart) GetAppVersion(
) string {
	return c.AppVersion
}

// Modify the chart by changing only the version
func (c *Chart) WithVersion(
	version string,
) *Chart { 
	c.Version = version
	return c
}

// Modify the chart by changing only the appVersion
func (c *Chart) WithAppVersion(
	appVersion string,
) *Chart { 
	c.AppVersion = appVersion
	return c
}

// Modify the chart by appending a suffix to the original version
func (c *Chart) WithRandomOriginalVersionSuffix(
) *Chart {
	return c.WithOriginalVersionSuffix(RandomString(randomSuffixLength))
}

// Modify the chart by appending a suffix to the original version
func (c *Chart) WithOriginalVersionSuffix(
	suffix string,
) *Chart {
	versionWithSuffix := fmt.Sprintf("%s-%s", originalVersion, suffix)
	return c.WithVersion(versionWithSuffix)
}

// Reset the chart to its original content
func (c *Chart) Reset(
	ctx context.Context,
) *Chart {
	c.DaggerDirectory.
		WithoutFile("Chart.yaml").
		WithNewFile("Chart.yaml", c.originalContent).
		Sync(ctx)
	return c
}

// Sync the directory with changes
func (c *Chart) Sync(
	ctx context.Context,
) *Chart {
	var buf bytes.Buffer
	err := c.ContentTemplate.ExecuteTemplate(&buf, c.Name, c)
	if err != nil {
		panic(err)
	}

	c.DaggerDirectory, err = c.DaggerDirectory.
		WithoutFile("Chart.yaml").
		WithNewFile("Chart.yaml", buf.String()).
		Sync(ctx)

	if err != nil {
		panic(err)
	}

	return c
}

func newChart(
	ctx context.Context,
	name string,
	directory string,
	templateString string,
	version string,
	appVersion string,
	sessionKey string,
) *Chart {
	ourSessionKey := sessionKey
	if (ourSessionKey == "") {
		ourSessionKey = RandomString(10)
	}
	ourName := fmt.Sprintf("%s-%s", name, ourSessionKey)
	c := Chart{
		Name: ourName,
		Directory: directory,
		ContentTemplate: template.Must(template.New(ourName).Parse(templateString)),
		SessionKey: ourSessionKey,
		Version: version,
		AppVersion: appVersion,
	}

	c.DaggerDirectory = dag.CurrentModule().Source().Directory(c.Directory)
	return c.Sync(ctx)
}

func NewMyChart(
	ctx context.Context,
) *Chart {
	return newChart(ctx, mychartName, mychartDir, mychartTemplateString, originalVersion, originalAppVersion, "")
}


func NewMyChartInSession(
	ctx context.Context,
	sessionKey string,
) *Chart {
	return newChart(ctx, mychartName, mychartDir, mychartTemplateString, originalVersion, originalAppVersion, sessionKey)
}

func NewMyDependentChart(
	ctx context.Context,
) *Chart {
	return newChart(ctx, mydependentchartName, mydependentchartDir, mydependentchartTemplateString, originalVersion, originalAppVersion, "")
}

func NewMyDependentChartInSession(
	ctx context.Context,
	sessionKey string,
) *Chart {
	return newChart(ctx, mydependentchartName, mydependentchartDir, mydependentchartTemplateString, originalVersion, originalAppVersion, sessionKey)
}

// The payoff: A chart session that suffixes to the chart name a random SessionKey string for uniqueness of each test
func NewChartSession(
	ctx context.Context,
) (*Chart, *Chart) {
	Mychart := NewMyChart(ctx)
	Mydependentchart := NewMyDependentChartInSession(ctx, Mychart.SessionKey)
	return Mychart, Mydependentchart
}

func RandomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
