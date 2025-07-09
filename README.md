# Daggerverse Helm Module

[Dagger](https://dagger.io/) module for [daggerverse](https://daggerverse.dev/) providing [Helm](https://helm.sh/) functionality.

The Dagger module is located in the [helm](./helm/) directory.

## Usage

Basic usage guide.

The [helm](./helm/) directory contains a [daggerverse](https://daggerverse.dev/) [Dagger](https://dagger.io/) module.

Check the official Dagger Module documentation: https://docs.dagger.io/api/module-structure

The [Dagger CLI](https://docs.dagger.io/cli) is needed.

### Functions

List all functions of the module. This command is provided by the [Dagger CLI](https://docs.dagger.io/cli). 

```bash
dagger functions -m ./helm/
```

## Development

Basic development guide.

### Module Image

The Image used in the module is built by the [ci/](./ci/) Dagger module (publish function).

The [image-build](.github/workflows/image-build.yml) workflow builds and publishes the image.
This action can be triggered manually. It uses the GitHub secrets for the needed credentials.

### Setup/update Dagger module

```bash
dagger -m ./helm/ develop
```

## Contributors

Please add `gofmt -s -w .` to your `.git/hooks/pre-commit` hook.

## To Do

- [ ] Add more tools
- [ ] Add cache mounts
- [ ] Add environment variables
- [ ] Add more examples
