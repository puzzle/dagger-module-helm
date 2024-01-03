# daggerverse Helm Module

[Dagger](https://dagger.io/) module for [daggerverse](https://daggerverse.dev/) providing [Helm](https://helm.sh/)

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

### version

Get and display the version of the Helm Chart located at the directory given by the `--d` flag.

```bash
dagger call -m ./helm/ version --d mychart/
```

## development

Basic development guide.

### setup Dagger module

Setup the Dagger module.

Create the directory for the module and initialize it.

```bash
mkdir helm/
cd helm/

# initialize Dagger module
dagger mod init --sdk go --name helm
```

### setup development module

Setup the outer module to be able to develop the Dagger Helm module.

```bash
dagger mod init --sdk go --name modest
dagger mod use ./helm
```

Generate or re-generate the Go definitions file (dagger.gen.go) for use in code completion.

```bash
dagger mod install
```

The functions of the module are available by the `dag` variable. Type `dag.` in your Go file for code completion.


Update the module:

```bash
dagger mod update
```
