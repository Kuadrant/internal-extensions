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

## Build and Test

```bash
# Build
go build -o pipeline-policy .

# Test
go test ./...
```
