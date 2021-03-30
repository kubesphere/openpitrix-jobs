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

# Run go fmt against code
fmt:
	gofmt -w ./pkg ./cmd

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...



# find or download controller-gen
# download controller-gen if necessary
clientset: 
	./hack/generate_client.sh


openpitrix-jobs: fmt
	docker build . -t openpitrix-jobs:latest

