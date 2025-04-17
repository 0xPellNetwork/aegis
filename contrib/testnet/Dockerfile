ARG FAST_SYNC=false
ARG NODE_NAME=pellcored-local

FROM pellnetwork/pellnode:v1.0.1

ARG NODE_NAME
ARG FAST_SYNC

RUN apt-get update && apt-get install -y git

RUN wget https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2Fv1.5.0/cosmovisor-v1.5.0-linux-amd64.tar.gz \
    && tar -xzf cosmovisor-v1.5.0-linux-amd64.tar.gz -C /usr/local/bin/ \
    && rm cosmovisor-v1.5.0-linux-amd64.tar.gz

RUN pellcored init ${NODE_NAME}

RUN mkdir -p /root/.pellcored/cosmovisor/genesis/bin
RUN mkdir -p /root/.pellcored/cosmovisor/upgrades

RUN cp $(which pellcored) /root/.pellcored/cosmovisor/genesis/bin

COPY app.toml /root/.pellcored/config/app.toml
COPY client.toml /root/.pellcored/config/client.toml
COPY config.toml /root/.pellcored/config/config.toml
COPY genesis.json /root/.pellcored/config/genesis.json

ENV DAEMON_HOME=/root/.pellcored
ENV DAEMON_NAME=pellcored
ENV DAEMON_ALLOW_DOWNLOAD_BINARIES=false
ENV DAEMON_RESTART_AFTER_UPGRADE=true

CMD ["cosmovisor", "run", "start"]