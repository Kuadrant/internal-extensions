## Project Overview

This is a collection of internal/test Kuadrant extensions that dogfood the kuadrant-operator extension SDK (`pkg/extension/`). Each extension lives under `extensions/<name>/` and is built into a single container image.

## How Kuadrant Extensions Work

Extensions are separate Go binaries that communicate with the kuadrant-operator via gRPC over Unix domain sockets.

**Mounting into the operator:**
1. Build static Go binaries (one per extension)
2. Place each at `/extensions/<name>/<name>` in the operator container
3. On startup, the operator's extension manager scans `/extensions/`, discovers executables matching the `<dir>/<dir>` naming convention
4. The manager starts each extension as a child process, passing a Unix socket path as the first CLI argument
5. The extension connects back to the operator over that socket for CEL evaluation, data bindings, upstream registration, and pipeline commits

**SDK package:** `github.com/kuadrant/kuadrant-operator/pkg/extension/`
- `pkg/extension/controller` — controller builder (`NewBuilder()`)
- `pkg/extension/types` — action types (AllowAction, GRPCMethodAction, AddHeadersAction, WithResponseCodeAction), Pipeline interface, KuadrantCtx

**Reference extensions** (in-tree, at `cmd/extensions/` in kuadrant-operator):
- `threat-policy` — best reference for pipeline actions
- `oidc-policy`, `plan-policy`, `telemetry-policy` — other examples

## Project Structure

```
extensions/
  pipeline-policy/         # Generic PipelinePolicy extension
    main.go                # Entry point
    api/v1alpha1/          # CRD types, deepcopy, scheme registration
    internal/controller/   # Reconciler
config/
  crd/bases/               # Generated CRD YAMLs (make manifests)
  rbac/                    # ClusterRole and ClusterRoleBinding
examples/                  # Sample CRs
```

## Extensions

### pipeline-policy

A generic extension whose spec declaratively defines a full action pipeline — gRPC upstreams, request-phase actions (allow, grpc_method), and response-phase actions (add_headers, with_response_code). The reconciler reads the spec and translates it into SDK calls. No business logic — different scenarios are different YAML manifests.

**CRD:** `PipelinePolicy` (group: `extensions.kuadrant.io`)
- `targetRef` — Gateway API resource (HTTPRoute or Gateway)
- `actionMethods[]` — gRPC upstreams to register
- `request[]` — ordered request-phase actions
- `response[]` — ordered response-phase actions

## Build and Test

```bash
make build       # generate deepcopy + build all extensions to bin/
make manifests   # generate CRD YAMLs
make test        # run all tests
```

## Adding a New Extension

1. Create `extensions/<name>/` with `main.go`, `api/`, `internal/controller/`
2. Import paths use `github.com/crstrn13/internal-extensions/extensions/<name>/...`
3. Run `make build` — auto-discovers all extensions under `extensions/`
4. Run `make manifests` — generates CRDs for all extensions

## Deployment

The image is pushed to `quay.io/acristur/internal-extensions`. To deploy:
1. Install CRDs: `kubectl apply -f config/crd/bases/`
2. Install RBAC: `kubectl apply -f config/rbac/`
3. Mount the image's `/extensions/` directory into the kuadrant-operator pod
4. The operator discovers and starts all extensions automatically
