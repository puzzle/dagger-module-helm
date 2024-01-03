# Helm-Push

Usage

```bash
dagger mod init --name mymodule --sdk go
dagger call -m "github.com/schlapzz/dagger-modules/helm-push" package-push --registry registry.puzzle.ch --repository helm --username $REGISTRY_HELM_USER --password $REGISTRY_HELM_PASSWORD --d .
```
