FROM golang:1.13-alpine as builder

RUN apk add --no-cache git curl openssl

RUN mkdir -p /workspace/helm-package-repository/
WORKDIR /workspace/helm-package-repository/
COPY . .

RUN mkdir -p /release_bin
RUN CGO_ENABLED=0 GOBIN=/release_bin go install -mod=vendor -ldflags '-w -s'  kubesphere.io/openpitrix-jobs/cmd/import-app/...

RUN cd package && for pkg in $(cat urls.txt); do curl -O $pkg; done

FROM kubesphere/kubectl:v1.19.0

ARG HELM_VERSION=v3.5.2
ARG KUSTOMIZE_VERSION=v4.0.5

# install helm
RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    tar xvf helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    rm helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/bin/ && \
    rm -rf linux-amd64
