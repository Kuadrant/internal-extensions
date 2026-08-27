# Internal Extensions

An internal/test [Kuadrant](https://kuadrant.io) extension, [pipeline-policy](extensions/pipeline-policy/), built using the [kuadrant-operator extension SDK](https://github.com/Kuadrant/kuadrant-operator/tree/main/pkg/extension). Its `PipelinePolicy` CRD declaratively defines an action pipeline.

## Build

```bash
make build       # generate deepcopy + build the binary
make manifests   # generate the CRD YAML
make test        # run all tests
```

## Deploy

The extension runs standalone, as its own pod, connecting to the
kuadrant-operator's extension gRPC service over TCP.

```bash
# 1. Build the image locally
make docker-build

# To push to your own registry, override IMG:
# export IMG=your-registry/pipeline-policy:latest
# make docker-build
# make docker-push

# 2. Deploy CRDs, RBAC, and the extension (uses IMG to set the container image)
IMG=your-registry/pipeline-policy:latest make deploy

# Or apply resources manually:
# kubectl apply -k config/crd/bases/
# kubectl apply -f config/namespace.yaml
# kubectl apply -k config/rbac/
# kubectl apply -k config/deploy/

# 3. Create a policy
kubectl apply -f examples/policy.yml
```

