FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY main.go main.go
COPY api/ api/
COPY internal/ internal/

ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -o pipeline-policy .

FROM registry.access.redhat.com/ubi9-minimal:latest
WORKDIR /
COPY --from=builder /workspace/pipeline-policy /extensions/pipeline-policy/pipeline-policy
USER 65532:65532
