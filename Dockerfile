FROM golang:1.23 AS build-env

# Version to build. Default is empty
ARG VERSION

ARG LEDGER_ENABLED="false"
# Cosmos build options
ARG COSMOS_BUILD_OPTIONS=""
ARG BUILD_TAGS=""

# Install cli tools for building and final image
RUN apt-get update && apt-get install -y make git bash gcc curl jq

# Build
WORKDIR /go/delivery/pell-node
# First cache dependencies
ARG GITHUB_TOKEN
RUN if [ -z "$GITHUB_TOKEN" ]; then echo "GITHUB_TOKEN is not set" && exit 1; fi
RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/0xPellNetwork".insteadOf "https://github.com/0xPellNetwork"
COPY go.mod go.sum /go/delivery/pell-node/
RUN --mount=type=cache,target="/go/pkg/mod" go mod download

# Then copy everything else
COPY ./ /go/delivery/pell-node/
# If version is set, then checkout this version
RUN if [ -n "${VERSION}" ]; then \
    git fetch origin ${VERSION}; \
    git checkout -f ${VERSION}; \
    fi

RUN --mount=type=cache,target="/go/pkg/mod" \
    --mount=type=cache,target="/root/.cache/go-build" \
    LEDGER_ENABLED=$LEDGER_ENABLED \
    BUILD_TAGS=$BUILD_TAGS \
    COSMOS_BUILD_OPTIONS=$COSMOS_BUILD_OPTIONS \
    LINK_STATICALLY=false \
    make install

FROM debian:bookworm-slim AS run
RUN apt-get update && apt-get install -y bash curl jq wget openssh-server iproute2 git

# Install libraries
# Cosmwasm - Download correct libwasmvm version
COPY --from=build-env /go/delivery/pell-node/go.mod /tmp
RUN WASMVM_VERSION=$(grep github.com/CosmWasm/wasmvm /tmp/go.mod | cut -d' ' -f2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm.$(uname -m).so \
    -O /lib/libwasmvm.$(uname -m).so && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm.$(uname -m).so | grep $(cat /tmp/checksums.txt | grep libwasmvm.$(uname -m) | cut -d ' ' -f 1)
RUN rm -f /tmp/go.mod

COPY --from=build-env /go/bin/pellcored /usr/local/bin/pellcored
COPY --from=build-env /go/bin/pellclientd /usr/local/bin/pellclientd

COPY contrib/testnet/start-pellcored.sh /usr/local/bin/start-pellcored.sh

# Set the default shell
ENV SHELL /bin/bash
# Set home directory and user
WORKDIR /usr/local/bin
EXPOSE 26656
EXPOSE 1317
EXPOSE 8545
EXPOSE 8546
EXPOSE 9090
EXPOSE 26657
EXPOSE 9091

ENTRYPOINT ["/usr/local/bin/start-pellcored.sh"]