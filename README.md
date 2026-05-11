# PipelinePolicy

A standalone [Kuadrant](https://kuadrant.io) extension that provides a **PipelinePolicy** CRD. Unlike the built-in extensions that hardcode specific actions, PipelinePolicy is generic: its spec declaratively defines the full action pipeline — gRPC upstreams, request-phase actions, and response-phase actions. Different scenarios are different YAML manifests, no code changes needed.

Built using the [kuadrant-operator extension SDK](https://github.com/Kuadrant/kuadrant-operator/tree/main/pkg/extension).

## CRD Spec

| Field | Description |
|-------|-------------|
| `targetRef` | Gateway API resource (HTTPRoute or Gateway) |
| `actionMethods[]` | gRPC upstreams to register (name, url, service, method, messageTemplate) |
| `request[]` | Ordered request-phase actions — `allow` or `grpc_method` |
| `response[]` | Ordered response-phase actions — `add_headers` or `with_response_code` |

See [examples/policy.yml](examples/policy.yml) for a full example.

## Build

```bash
make build       # generate deepcopy + build binary
make manifests   # generate CRD YAML
make test        # run tests
```

## Deploy

```bash
# 1. Build and push the image
docker build . -t <registry>/pipeline-policy:latest
docker push <registry>/pipeline-policy:latest

# 2. Install the CRD and RBAC
kubectl apply -f config/crd/bases/
kubectl apply -f config/rbac/

# 3. Mount the binary into the kuadrant-operator
#    Add an init container to the operator deployment that copies the binary
#    into a shared volume at /extensions/pipeline-policy/pipeline-policy

# 4. Create a PipelinePolicy
kubectl apply -f examples/policy.yml
```

The operator's extension manager discovers the binary at `/extensions/pipeline-policy/pipeline-policy` and starts it as a child process.
