FROM kubesphere/kubectl:v1.19.0

ARG HELM_VERSION=v3.5.2
ARG KUSTOMIZE_VERSION=v4.0.5

# install helm
RUN wget https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    tar xvf helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    rm helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/bin/ && \
    rm -rf linux-amd64
