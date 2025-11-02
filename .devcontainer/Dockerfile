FROM mcr.microsoft.com/devcontainers/base:ubuntu

ARG USERNAME=vscode
ARG GOPATH=/go
ARG GOMODCACHE=$GOPATH/pkg/mod
ARG WORKSPACES_DIR=/workspaces

SHELL ["/bin/bash", "-c"]

# Base setup
ENV GOPATH=$GOPATH
ENV GOMODCACHE=$GOMODCACHE
ENV ASDF_DATA_DIR=/usr/local/asdf
ENV PATH=$ASDF_DATA_DIR/shims:$GOPATH/bin:$PATH

RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    build-essential \
    fzf \
    git \
    golang \
    shellcheck \
    shfmt \
    tree && \
    apt-get -y clean && \
    rm -rf /var/lib/apt/lists/* && \
    go install github.com/asdf-vm/asdf/cmd/asdf@latest && \
    echo ". <(asdf completion bash)" >> /etc/bash.bashrc

# Grant permissions to the user
RUN umask 0002 && \
    mkdir -p $GOPATH $ASDF_DATA_DIR $WORKSPACES_DIR && \
    groupadd -r golang && \
    groupadd -r asdf && \
    usermod -aG golang,asdf $USERNAME && \
    chown -R $USERNAME:golang $GOPATH && \
    chown -R $USERNAME:asdf $ASDF_DATA_DIR && \
    chown -R $USERNAME:$USERNAME $WORKSPACES_DIR && \
    find $GOPATH $ASDF_DATA_DIR $WORKSPACES_DIR -type d -exec chmod g+s {} +

# Configure git
RUN git config --system init.defaultBranch main && \
    git config --system alias.cof "!git branch -a | fzf | xargs git checkout"

USER $USERNAME
WORKDIR $WORKSPACES_DIR
