## Project Overview

This repo hosts a single internal/test Kuadrant extension, pipeline-policy, that dogfoods the kuadrant-operator extension SDK (`pkg/extension/`). It lives under `extensions/pipeline-policy/`, built into its own container image and deployed as its own pod.

## How Kuadrant Extensions Work

Extensions are separate Go binaries, deployed as their own pods, that communicate with the kuadrant-operator via gRPC over TCP.

**Deployment:**
1. Build a static Go binary for the extension
2. Deploy it as its own pod with `KUADRANT_EXTENSION_ADDRESS` (pointing at the operator's `kuadrant-operator-extensions` Service on port 50052) and `KUADRANT_EXTENSION_TOKEN_FILE` (path to a projected ServiceAccount token) env vars
3. The extension authenticates to the operator using a projected SA token (audience `kuadrant-extensions`); the operator validates it via TokenReview and checks RBAC for the `register` verb on the `policyregistrations` virtual resource
4. The extension connects to the operator over that address for CEL evaluation, data bindings, upstream registration, and pipeline commits

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
  namespace.yaml           # Shared namespace (prerequisite for RBAC and deploy)
  crd/bases/               # Generated CRD YAMLs (make manifests)
  rbac/                    # ClusterRole, ClusterRoleBinding, and the extension's own ServiceAccount
  deploy/                  # Deployment for running the extension as its own pod
examples/                  # Sample CRs
```

## pipeline-policy

A generic extension whose spec declaratively defines a full action pipeline — gRPC upstreams, request-phase actions (allow, grpc_method), and response-phase actions (add_headers, with_response_code). The reconciler reads the spec and translates it into SDK calls. No business logic — different scenarios are different YAML manifests.

**CRD:** `PipelinePolicy` (group: `extensions.kuadrant.io`)
- `targetRef` — Gateway API resource (HTTPRoute or Gateway)
- `actionMethods[]` — gRPC upstreams to register
- `request[]` — ordered request-phase actions
- `response[]` — ordered response-phase actions

## Build and Test

```bash
make build       # generate deepcopy + build the binary to bin/
make manifests   # generate the CRD YAML
make test        # run all tests
```

## Deployment

The extension runs as its own pod, authenticating to the kuadrant-operator
via a projected ServiceAccount token (audience `kuadrant-extensions`). The
operator validates the token via TokenReview and authorizes the extension
through RBAC on the `policyregistrations` virtual resource.

1. Build and push the image: `make docker-build` and `make docker-push`
2. Install CRDs and the namespace: `kubectl apply -f config/crd/bases/` and
   `kubectl apply -f config/namespace.yaml` — the namespace must exist
   first since `config/rbac/serviceaccount.yaml` targets it
3. Install RBAC: `kubectl apply -f config/rbac/` (ClusterRole, ClusterRoleBinding,
   and ServiceAccount)
4. Deploy the extension: `kubectl apply -k config/deploy/` (Namespace, Deployment)

Files:
- `Dockerfile` — builds only the pipeline-policy binary (doesn't rely on the
  operator image or a shared `/extensions` filesystem convention)
- `config/deploy/` — Namespace and Deployment for running the extension
  as its own pod
- `config/rbac/` — ClusterRole (including `register` verb on `policyregistrations`
  for `PipelinePolicy`), ClusterRoleBinding, and ServiceAccount
