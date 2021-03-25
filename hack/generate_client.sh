#!/bin/bash

set -e

GV="application:v1alpha1"

rm -rf ./pkg/client
./hack/generate_group.sh "client,lister,informer" kubesphere.io/openpitrix-jobs/pkg/client kubesphere.io/openpitrix-jobs/pkg/apis "$GV" --output-base=./  -h "$PWD/hack/boilerplate.go.txt"
mv kubesphere.io/openpitrix-jobs/pkg/client ./pkg/
rm -rf ./kubesphere.io
