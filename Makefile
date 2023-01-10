# Image URL to use all building/pushing image targets
AGENTIMG ?= lcw2/octopus-agent:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1,generateEmbeddedObjectMeta=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif




# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./..." output:crd:artifacts:config=deploy/crds/


.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: update-codegen
update-codegen:  ## generetor clientset informer inderx code
	bash ./hack/update-codegen.sh

.PHONY: docker-build-agent
docker-build-agent:
	docker build --network host -f ./build/agent/Dockerfile -t ${AGENTIMG} .

# find or download controller-gen
# download controller-gen if necessary

controller-gen:
ifeq (, $(shell which controller-gen))
	echo "Start to install controller-gen tool"
	export GO111MODULE=on
	go install -v sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2
	export GOPATH="${GOPATH:-$(go env GOPATH)}"
	export PATH=$PATH:$GOPATH/bin
CONTROLLER_GEN=$(shell which controller-gen)
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

