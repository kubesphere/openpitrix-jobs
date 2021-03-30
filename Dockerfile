FROM golang:1.13-alpine as builder

RUN apk add --no-cache git curl openssl

RUN mkdir -p /workspace/helm-package-repository/
WORKDIR /workspace/helm-package-repository/
COPY . .

RUN mkdir -p /release_bin
RUN CGO_ENABLED=0 GOBIN=/release_bin go install -mod=vendor -ldflags '-w -s'  kubesphere.io/openpitrix-jobs/cmd/import-app/...

FROM kubesphere/kubectl:v1.19.0

WORKDIR /root
COPY urls.txt /root
RUN mkdir -p package && cp urls.txt package  && cd /root/package && for pkg in $(cat urls.txt); do curl -O $pkg; done

COPY --from=builder /release_bin/* /usr/local/bin/
COPY start.sh /root/

