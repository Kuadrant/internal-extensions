# Builds an image containing only the pipeline-policy binary, for deployment
# as its own pod (see config/deploy/). The extension connects to the
# kuadrant-operator over TCP, so this doesn't need anything from the
# operator image itself.
FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY extensions/pipeline-policy/ extensions/pipeline-policy/

ARG TARGETARCH

RUN mkdir -p /workspace/bin && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -o /workspace/bin/pipeline-policy ./extensions/pipeline-policy

FROM registry.access.redhat.com/ubi9-minimal:latest
WORKDIR /

COPY --from=builder /workspace/bin/pipeline-policy /pipeline-policy

USER 65532:65532

ENTRYPOINT ["/pipeline-policy"]
