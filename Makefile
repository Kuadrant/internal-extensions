CONTROLLER_GEN_VERSION ?= v0.19.0
CONTROLLER_GEN ?= $(shell which controller-gen 2>/dev/null)

EXTENSIONS := $(shell ls extensions/)

.PHONY: generate
generate: controller-gen
	@for ext in $(EXTENSIONS); do \
		echo "Generating deepcopy for $$ext ..."; \
		$(CONTROLLER_GEN) object paths=./extensions/$$ext/api/...; \
	done

.PHONY: manifests
manifests: controller-gen
	@for ext in $(EXTENSIONS); do \
		echo "Generating CRD manifests for $$ext ..."; \
		$(CONTROLLER_GEN) crd paths=./extensions/$$ext/api/... output:crd:dir=config/crd/bases; \
	done

.PHONY: build
build: generate
	@for ext in $(EXTENSIONS); do \
		echo "Building $$ext ..."; \
		go build -o bin/$$ext ./extensions/$$ext; \
	done

.PHONY: test
test:
	go test ./...

.PHONY: controller-gen
controller-gen:
ifeq (,$(CONTROLLER_GEN))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)
endif
