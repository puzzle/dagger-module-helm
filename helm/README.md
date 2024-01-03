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

### version

Get and display the version of the Helm Chart located at the directory given by the `--directory` flag.

Usage

```bash
dagger call -m "github.com/puzzle/dagger-module-helm/helm" version --directory ./mychart/
```
