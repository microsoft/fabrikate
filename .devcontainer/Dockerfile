FROM ubuntu:bionic

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG GO_VERSION=1.14.2
ARG GOLANGCI_LINT_VERSION=1.25.0
ARG HELM_VERSION=3.1.2
ARG DOCKER_VERSION=19.03.8
ARG DOCKER_LOCK_VERSION=0.2.0

RUN apt-get update && \
    apt-get -y install vim curl wget gcc software-properties-common && \
    add-apt-repository ppa:git-core/ppa -y && \
    apt-get update && \
	apt-get install git -y && \
    curl -fsSL "https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz" | tar -xz && \
    mv linux-amd64/helm /usr/local/bin && \
    rm -rf linux-amd64 && \
    curl -fsSL "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" | tar -xz -C /usr/local docker/docker && \
    mkdir -p "${HOME}/.docker/cli-plugins" && \
    curl -fsSL "https://github.com/michaelperel/docker-lock/releases/download/v${DOCKER_LOCK_VERSION}/docker-lock-linux" -o "${HOME}/.docker/cli-plugins/docker-lock" && \
    chmod +x "${HOME}/.docker/cli-plugins/docker-lock" && \
    curl -fsSL "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" | tar -xz -C /usr/local && \
    curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b "${HOME}/go/bin" "v${GOLANGCI_LINT_VERSION}" && \
    echo "export PATH=/usr/local/go/bin:${HOME}/go/bin:/usr/local/docker:${PATH}" >> "${HOME}/.bashrc" && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*