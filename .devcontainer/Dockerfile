FROM ubuntu:bionic

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG GO_VERSION=1.14.2
ARG GOLANGCI_LINT_VERSION=1.25.0
RUN apt-get update && \
    apt-get -y install vim curl wget gcc software-properties-common && \
    add-apt-repository ppa:git-core/ppa -y && \
    apt-get update && \
	apt-get install git -y && \
    curl -fsSL https://get.docker.com | bash && \
    curl -fsSL "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" | tar -xz -C /usr/local && \
    curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b "${HOME}/go/bin" "v${GOLANGCI_LINT_VERSION}" && \
    echo "export PATH=/usr/local/go/bin:${HOME}/go/bin:${PATH}" >> "${HOME}/.bashrc" && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
