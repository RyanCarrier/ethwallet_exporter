# ethwallet_exporter

Ethereum and token balance exporter for prometheus

## Run

Go

```
go mod download
go build -o ./ethwallet_exporter
./ethwallet_exporter --geth="http://geth.rpc.endpoint" --addresses="address1.eth,0xhexofwallet"
```

Docker

```
docker pull ryancarrier/ethwallet_exporter
docker run -it --rm ryancarrier/ethwallet_exporter --geth="http://geth.rpc.endpoint" --addresses="address1.eth,0xhexofwallet"
```



## TODO

add 9887 port to prometheus default port allocations

better readme (unlikely)


