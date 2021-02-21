# poison-pill-manager

Poison Pill Manager is an integration of poison-pill operator with OLM.

# CRD

Currenly the CRD contains nothing. It's a placholder for future configuration:
```yaml
apiVersion: config.poison-pill.io.poison-pill.io/v1alpha1
kind: PoisonPillConfig
metadata:
  name: poisonpillconfig-sample
```
# Dev

## Build & Push:
```bash
make docker-build IMG=quay.io/<user>/poison-pill-manager:latest
make docker-push IMG=quay.io/<user>/poison-pill-manager:latest
```

## Create bundle image
```bash
make bundle-build BUNDLE_IMG=quay.io/<user>/poison-pill-manager-bundle:latest
make docker-push IMG=quay.io/<user>/poison-pill-manager-bundle:latest
```

## Create index image
```bash
opm index add --bundles quay.io/<user>/poison-pill-manager:latest --tag quay.io/<user>/poison-pill-manager-index:latest
docker push quay.io/<user>/poison-pill-manager-index:latest
```

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: ppill-manifest
  namespace: default
spec:
  sourceType: grpc
  image: quay.io/<user>/poison-pill-manager-index:latest
```

To deploy the operator in a test cluster you'll also need `OperatorGroup` and `Subscription`:
```yaml
apiVersion: operators.coreos.com/v1alpha2
kind: OperatorGroup
metadata:
  name: my-ppill-group
  namespace: default
spec:
```
```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: ppill-subscription
  namespace: default 
spec:
  channel: alpha
  name: poison-pill-manager
  source: ppill-manifest
  sourceNamespace: default
```





