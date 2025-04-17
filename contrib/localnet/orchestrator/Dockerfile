ARG TAG=latest
FROM pellnode:${TAG} AS pell
FROM ethereum/client-go:v1.10.26 AS geth
FROM golang:1.23 AS orchestrator

RUN apt-get update && apt-get install -y bash curl jq wget openssh-server
RUN go install github.com/mikefarah/yq/v4@v4.44.3

COPY --from=pell /lib/libwasmvm.* /lib/
COPY --from=pell /usr/local/bin/pellcored /go/bin/
COPY --from=pell /usr/local/bin/pellclientd /go/bin/
COPY --from=pell /usr/local/bin/pelle2e /go/bin/
COPY --from=geth /usr/local/bin/geth /go/bin/

COPY contrib/localnet/ssh /root/.ssh
RUN chmod 700 /root/.ssh
RUN chmod 600 /root/.ssh/*

COPY contrib/localnet/orchestrator/*.sh /work/

RUN chmod +x /work/*.sh

ENV GOPATH /go
ENV GOOS=linux
ENV CGO_ENABLED=1
WORKDIR /go/delivery/pell-node

RUN chmod +x /go/bin/pelle2e

WORKDIR /work
