#!/bin/bash

set -e

GV="application:v1alpha1"

rm -rf ./pkg/client
./hack/generate_group.sh "client,lister,informer" github.com/xyz-li/openpitrix-job/pkg/client github.com/xyz-li/openpitrix-job/pkg/apis "$GV" --output-base=./  -h "$PWD/hack/boilerplate.go.txt"
mv github.com/xyz-li/openpitrix-job/pkg/client ./pkg/
rm -rf ./kubesphere.io
