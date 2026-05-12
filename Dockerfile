FROM quay.io/kuadrant/kuadrant-operator:latest AS operator

FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY extensions/ extensions/

ARG TARGETARCH

RUN for ext in extensions/*/; do \
        name=$(basename "$ext"); \
        echo "Building $name ..."; \
        CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -o "bin/$name" "./extensions/$name"; \
    done

FROM registry.access.redhat.com/ubi9-minimal:latest
WORKDIR /

COPY --from=operator /extensions/ /extensions/

COPY --from=builder /workspace/bin/ /tmp/bins/
RUN for bin in /tmp/bins/*; do \
        name=$(basename "$bin"); \
        mkdir -p "/extensions/$name"; \
        cp "$bin" "/extensions/$name/$name"; \
    done && rm -rf /tmp/bins
USER 65532:65532
