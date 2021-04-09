# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

# Run go fmt against code
fmt:
	go fmt ./...
	cd cmd/dump-all && go fmt ./...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# find or download controller-gen
# download controller-gen if necessary
clientset: 
	./hack/generate_client.sh


openpitrix-jobs: fmt
	docker build . -t openpitrix-jobs:latest --build-arg BUILDDATE=$$(date +%s)

mod-vendor:
	go mod vendor
	cd cmd/dump-all && go mod vendor