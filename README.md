# Internal Extensions

A collection of internal/test [Kuadrant](https://kuadrant.io) extensions built using the [kuadrant-operator extension SDK](https://github.com/Kuadrant/kuadrant-operator/tree/main/pkg/extension).

## Extensions

| Extension | Description |
|-----------|-------------|
| [pipeline-policy](extensions/pipeline-policy/) | Generic PipelinePolicy CRD that declaratively defines action pipelines |

## Build

```bash
make build       # generate deepcopy + build all extensions
make manifests   # generate CRD YAMLs
make test        # run all tests
```

## Deploy

```bash
# 1. Build and push the image (contains all extension binaries)
docker build . -t quay.io/acristur/internal-extensions:latest
docker push quay.io/acristur/internal-extensions:latest

# 2. Install CRDs and RBAC
kubectl apply -f config/crd/bases/
kubectl apply -f config/rbac/

# 3. Mount the binaries into the kuadrant-operator
#    The image places each extension at /extensions/<name>/<name>
#    The operator's extension manager discovers and starts them automatically

# 4. Create a policy
kubectl apply -f examples/policy.yml
```

## Adding a New Extension

1. Create `extensions/<name>/` with `main.go`, `api/`, and `internal/controller/`
2. Follow the same patterns as `extensions/pipeline-policy/`
3. Run `make build` — it auto-discovers and builds all extensions
