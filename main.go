package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

const ethipc = "/eth.ipc"

var (
	addressList     []Address
	balanceList     []Balance
	port            string
	url             string
	lastRefresh     time.Duration
	eth             *ethclient.Client
	refreshDuration time.Duration = time.Second * 15
)

func init() {
	var err error
	portCustom := os.Getenv("PORT")
	duration := os.Getenv("DURATION")
	url = os.Getenv("GETH")
	addresses := os.Getenv("ADDRESSES")
	if len(portCustom) != 0 {
		port = portCustom
	}
	if len(duration) != 0 {
		refreshDuration, err = time.ParseDuration(duration)
		if err != nil {
			panic(err)
		}
	}
	if len(url) == 0 {
		if _, err = os.Stat(ethipc); err != nil {
			panic("no geth url supplied (or ipc file in /geth.ipc)")
		}
		url = ethipc
	}
	if len(addresses) == 0 {
		panic("no addresses supplied")
	}
	parseAddresses(addresses)
}

func init() {
	var err error
	eth, err = ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
}

func main() {
	go walletLoop()
	time.Tick(refreshDuration)
	http.HandleFunc("/metrics", handleMetrics)
	panic(http.ListenAndServe("0.0.0.0:"+port, nil))
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	var resp []string
	for _, v := range balanceList {
		resp = append(resp, fmt.Sprintf("crypto_balance{name=\"%v\",address=\"%v\"} %v", v.name, v.address, v.balance))
	}
	resp = append(resp, fmt.Sprintf("crypto_load_seconds %0.2f", lastRefresh.Seconds()))
	fmt.Fprintln(w, strings.Join(resp, "\n"))
}
