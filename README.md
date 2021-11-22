# ethwallet_exporter

Ethereum and token balance exporter for prometheus

| Environment variable | Default | Use |
| --- | --- | --- |
|PORT|8084| port to host webserever on (and listen for /metrics)
|DURATION|15s|Duration between account balance refreshes|
|GETH|localhost:8545|eth RPC|
|ADDRESSES|nil (REQUIRED)|addresses to scan, hex or ENS|
|CACHE|0| How many refreshes until a manual rescan for unknown token balances|

## TODO

test docker

change to flags instead of env

better readme (unlikely)


