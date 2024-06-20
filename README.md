# daggerverse Helm Module

[Dagger](https://dagger.io/) module for [daggerverse](https://daggerverse.dev/) providing [Helm](https://helm.sh/) functionality.

The Dagger module is located in the [helm](./helm/) directory.

## usage

Basic usage guide.

The [helm](./helm/) directory contains a [daggerverse](https://daggerverse.dev/) [Dagger](https://dagger.io/) module.

Check the official Dagger Module documentation: https://docs.dagger.io/zenith/

The [Dagger CLI](https://docs.dagger.io/cli) is needed.

### functions

List all functions of the module. This command is provided by the [Dagger CLI](https://docs.dagger.io/cli). 

```bash
dagger functions -m ./helm/
```

The helm module is referenced locally.

See the module [readme](./helm/README.md) or the method comments for more details.

## development

Basic development guide.

### setup Dagger module

Setup the Dagger module.

Create the directory for the module and initialize it.

```bash
mkdir helm/
cd helm/

# initialize Dagger module
dagger init
dagger develop --sdk go --source helm
```

## To Do

- [ ] Add more tools
- [ ] Add cache mounts
- [ ] Add environment variables
- [ ] Add more examples
- [ ] Add tests
