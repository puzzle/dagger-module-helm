# daggerverse Helm Module

[Dagger](https://dagger.io/) module for [daggerverse](https://daggerverse.dev/) providing [Helm](https://helm.sh/)

The Dagger module is located in the [helm](./helm/) directory.

## usage

Basic usage guide.

The [helm](./helm/) directory contains a [daggerverse](https://daggerverse.dev/) [Dagger](https://dagger.io/) module.

Check the official Dagger Module documentation: https://docs.dagger.io/zenith/

Run all commands from the [helm](./helm/) directory. The [Dagger CLI](https://docs.dagger.io/cli) is needed.

### functions

List all functions of the module. This command is provided by the [Dagger CLI](https://docs.dagger.io/cli). 

```bash
dagger functions
```

### version

Get and display the version of the Helm Chart located at the directory given by the `--d` flag.

```bash
dagger call version --d mychart/
```

## development

Basic development guide.

Run all commands from the [helm](./helm/) directory.
