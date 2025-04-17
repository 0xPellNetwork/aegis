
set -x
wait_for_contracts_deployment() {
  max_tries=50
  intervals=5
  tries=0
  filepath="../../deployments/localhost/MockDVSRegistryRouter-Proxy.json"

  while [ $tries -lt $max_tries ]; do
    ret=$(ssh hardhat "if [ -f $filepath ]; then echo 0; else echo 1; fi" 2>/dev/null || echo 1)
    if [ $ret -eq 0 ]; then
      echo "$filepath found. Deployment is successful."
      return 0
    fi
    tries=$((tries + 1))
    sleep $intervals
  done

  echo "Contracts not deployed after $max_tries tries"
  exit 1
}


# /usr/sbin/sshd

## Run geth in the background
nohup geth --dev --http --http.addr 0.0.0.0 --http.vhosts "*" --http.port 8545 --http.api "eth,web3,net,debug" --gcmode archive  --datadir /app/data &
while true; do
    geth --exec 'eth.sendTransaction({from: eth.coinbase, to: eth.coinbase, value: web3.toWei(1,"ether")})' attach http://127.0.0.1:8545
    sleep 2
done