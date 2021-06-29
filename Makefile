DOCKER_COMMAND ?= sudo docker

DOCKER_IMAGE ?= quay.io/erdii/configmap-puller:dev

build:
	$(DOCKER_COMMAND) build -t $(DOCKER_IMAGE) .
.PHONY: build

push:
	$(DOCKER_COMMAND) push $(DOCKER_IMAGE)
.PHONY: push

KIND_KUBECONFIG := kubeconfig
$(KIND_KUBECONFIG):
	source hack/determine-container-runtime.sh && \
		$$KIND_COMMAND create cluster --name foo --kubeconfig $(KIND_KUBECONFIG) --config kind.yaml && \
		if [[ ! -O "$(KIND_KUBECONFIG)" ]]; then \
			sudo chown $$USER: "$(KIND_KUBECONFIG)"; \
		fi;

rm-kind:
	source hack/determine-container-runtime.sh && \
		$$KIND_COMMAND delete cluster --name foo --kubeconfig $(KIND_KUBECONFIG) && \
			rm -f "$(KIND_KUBECONFIG)"
.PHONY: rm-kind
