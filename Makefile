# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OUTPUT_DIR=bin
GOFLAGS=-mod=mod
define ALL_HELP_INFO
# Build code.
#
# Args:
#   WHAT: Directory names to build.  If any of these directories has a 'main'
#     package, the build will produce executable files under $(OUT_DIR).
#     If not specified, "everything" will be built.
#   GOFLAGS: Extra flags to pass to 'go' when building.
#   GOLDFLAGS: Extra linking flags passed to 'go' when building.
#   GOGCFLAGS: Additional go compile flags passed to 'go' when building.
#
# Example:
#   make
#   make all
#   make all WHAT=cmd/ks-apiserver
#     Note: Use the -N -l options to disable compiler optimizations an inlining.
#           Using these build options allows you to subsequently use source
#           debugging tools like delve.
endef


# Build e2e binary
e2e: fmt vet
	hack/build_e2e.sh test/e2e

# Run go fmt against code
fmt:
	gofmt -w ./pkg ./cmd ./tools ./api

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...



# find or download controller-gen
# download controller-gen if necessary
clientset: 
	./hack/generate_client.sh


# Currently in the upgrade phase of controller tools.
# But the new controller tools are not compatible with the old version.
# With these commands you may need to manually modify the generated code
# So don't use it unless you know it very deeply
internal-crds:
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./pkg/apis/network/..." output:crd:artifacts:config=config/crd/bases

internal-generate-apis: internal-controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/apis/network/...

internal-controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.0-beta.4
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

network-rbac:
	$(CONTROLLER_GEN) paths=./pkg/controller/network/provider/ paths=./pkg/controller/network/ rbac:roleName=network-manager output:rbac:artifacts:config=kustomize/network/calico-k8s
	$(CONTROLLER_GEN) paths=./pkg/controller/network/ rbac:roleName=network-manager output:rbac:artifacts:config=kustomize/network/calico-etcd
