CONTROLLER_GEN_VERSION ?= v0.19.0
CONTROLLER_GEN ?= $(shell which controller-gen 2>/dev/null)

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object paths=./extensions/pipeline-policy/api/...

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) crd paths=./extensions/pipeline-policy/api/... output:crd:dir=config/crd/bases

.PHONY: build
build: generate
	mkdir -p bin
	go build -o bin/pipeline-policy ./extensions/pipeline-policy

.PHONY: test
test:
	go test ./...

IMG ?= pipeline-policy:dev
PLATFORMS ?= linux/amd64
CONTAINER_TOOL ?= $(shell which podman 2>/dev/null || which docker 2>/dev/null)

.PHONY: docker-build
docker-build:
	$(CONTAINER_TOOL) build --platform $(PLATFORMS) -t $(IMG) .

.PHONY: docker-push
docker-push:
	$(CONTAINER_TOOL) push $(IMG)

.PHONY: deploy
deploy:
	kubectl apply -k config/crd/bases/
	kubectl apply -f config/namespace.yaml
	kubectl apply -k config/rbac/
	kubectl apply -k config/deploy/
	kubectl set image deployment/pipeline-policy-extension pipeline-policy=$(IMG) -n kuadrant-extensions

.PHONY: controller-gen
controller-gen:
ifeq (,$(CONTROLLER_GEN))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)
endif
