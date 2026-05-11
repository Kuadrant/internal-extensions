## Project Overview

This is a standalone Kuadrant extension that dogfoods the kuadrant-operator extension SDK (`pkg/extension/`). It exists in its own repository to validate the out-of-tree developer experience for building extensions.

The extension provides a **PipelinePolicy** CRD whose spec declaratively defines a full action pipeline — gRPC upstreams, request-phase actions (allow, grpc_method), and response-phase actions (add_headers, with_response_code). The reconciler is generic: it reads the spec and translates it into SDK calls. No business logic — different scenarios are different YAML manifests.

## How Kuadrant Extensions Work

Extensions are separate Go binaries that communicate with the kuadrant-operator via gRPC over Unix domain sockets.

**Mounting into the operator:**
1. Build a static Go binary
2. Place it at `/extensions/<name>/<name>` in the operator container (e.g., `/extensions/pipeline-policy/pipeline-policy`)
3. On startup, the operator's extension manager scans `/extensions/`, discovers executables matching the `<dir>/<dir>` naming convention
4. The manager starts each extension as a child process, passing a Unix socket path as the first CLI argument
5. The extension connects back to the operator over that socket for CEL evaluation, data bindings, upstream registration, and pipeline commits

**SDK package:** `github.com/kuadrant/kuadrant-operator/pkg/extension/`
- `pkg/extension/controller` — controller builder (`NewBuilder()`)
- `pkg/extension/types` — action types (AllowAction, GRPCMethodAction, AddHeadersAction, WithResponseCodeAction), Pipeline interface, KuadrantCtx

**Reference extensions** (in-tree, at `cmd/extensions/` in kuadrant-operator):
- `threat-policy` — best reference for pipeline actions
- `oidc-policy`, `plan-policy`, `telemetry-policy` — other examples

## CRD Design

The PipelinePolicy spec contains:
- `targetRef` — Gateway API resource (HTTPRoute or Gateway)
- `actionMethods[]` — gRPC upstreams to register (name, url, service, method, messageTemplate)
- `request[]` — ordered request-phase actions (type, predicate, intention, method, var)
- `response[]` — ordered response-phase actions (type, predicate, headersToAdd, responseCode)

## Project Structure

- `main.go` — entry point: scheme registration, builder, controller start
- `api/v1alpha1/` — CRD types, deepcopy, scheme registration (group: `extensions.kuadrant.io`)
- `internal/controller/` — generic reconciler that iterates spec arrays and maps to SDK calls
- `config/crd/bases/` — generated CRD YAML (`make manifests`)
- `config/rbac/` — ClusterRole and ClusterRoleBinding for the operator service account
- `examples/` — sample PipelinePolicy CR

## Build and Test

```bash
make build       # generate deepcopy + build binary
make manifests   # generate CRD YAML
make test        # run tests
```

## Deployment

The extension image is pushed to `quay.io/acristur/pipeline-policy`. To deploy:
1. Install CRD: `kubectl apply -f config/crd/bases/`
2. Install RBAC: `kubectl apply -f config/rbac/`
3. Mount the binary into the kuadrant-operator pod at `/extensions/pipeline-policy/pipeline-policy`
4. The operator discovers and starts it automatically
