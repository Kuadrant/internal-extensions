CONTROLLER_GEN_VERSION ?= v0.19.0
CONTROLLER_GEN ?= $(shell which controller-gen 2>/dev/null)

.PHONY: generate
generate: controller-gen ## Generate DeepCopy methods.
	$(CONTROLLER_GEN) object paths=./api/...

.PHONY: manifests
manifests: controller-gen ## Generate CRD manifests.
	$(CONTROLLER_GEN) crd paths=./api/... output:crd:dir=config/crd/bases

.PHONY: build
build: generate ## Build the extension binary.
	go build -o pipeline-policy .

.PHONY: test
test: ## Run tests.
	go test ./...

.PHONY: controller-gen
controller-gen: ## Install controller-gen if not present.
ifeq (,$(CONTROLLER_GEN))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)
endif
