FROM golang:1.13-alpine as builder

RUN apk add --no-cache git curl openssl

RUN mkdir -p /workspace/helm-package-repository/
WORKDIR /workspace/helm-package-repository/
COPY . .

RUN mkdir -p /release_bin
RUN CGO_ENABLED=0 GOBIN=/release_bin go install -mod=vendor -ldflags '-w -s'  kubesphere.io/openpitrix-jobs/cmd/import-app/...
RUN CGO_ENABLED=0 GOBIN=/release_bin go install -mod=vendor -ldflags '-w -s'  kubesphere.io/openpitrix-jobs/cmd/start-jobs/...

FROM 139.198.9.238:30002/library/dump-all as dump
#FROM openpitrix/dump-all:latest as dump

FROM alpine:3.7
RUN apk add --update ca-certificates && update-ca-certificates

WORKDIR /root
COPY urls.txt /root
RUN mkdir -p package && cp urls.txt package  && cd /root/package && for pkg in $(cat urls.txt); do wget $pkg; done

COPY --from=dump /usr/local/bin/dump-all /usr/local/bin
COPY --from=builder /release_bin/* /usr/local/bin/
