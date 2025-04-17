ARG TAG=latest
FROM ethereum/client-go:v1.10.26 AS geth
FROM pellnode:${TAG} AS pell

COPY --from=geth /usr/local/bin/geth /usr/local/bin/
COPY contrib/localnet/ssh /root/.ssh
RUN chmod 700 /root/.ssh
RUN chmod 600 /root/.ssh/*

WORKDIR /app
COPY contrib/localnet/geth/start-eth.sh /app/start-eth.sh

ENTRYPOINT [ "bash" ]
CMD [ "start-eth.sh" ]
EXPOSE 8545