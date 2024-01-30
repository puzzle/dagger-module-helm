# Helm

Module Home: https://github.com/puzzle/dagger-module-helm

Provides [Helm](https://helm.sh/) functionality as [Dagger Module](https://daggerverse.dev/).

## Functions

### packagePush

Usage

```bash
dagger call -m "github.com/puzzle/dagger-module-helm/helm" packagePush \
  --registry registry.puzzle.ch \
  --repository helm \
  --username $REGISTRY_HELM_USER \
  --password $REGISTRY_HELM_PASSWORD \
  --directory ./mychart/
```

### test

Run [Helm unittests](https://github.com/helm-unittest/helm-unittest) with the given directory and files.

Usage

```bash
dagger call -m "github.com/puzzle/dagger-module-helm/helm" test --directory ./mychart/ --args "."
```

Provide the helm chart directory with pointing to it with the `--directory` flag. Add the directory location with `"."` as `--args` parameter to tell helm unittest where to find the helm chart with the tests.

### version

Get and display the version of the Helm Chart located at the directory given by the `--directory` flag.

Usage

```bash
dagger call -m "github.com/puzzle/dagger-module-helm/helm" version --directory ./mychart/
```
