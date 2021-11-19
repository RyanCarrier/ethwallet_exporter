package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	addressList     []Address = make([]Address, 0)
	balanceList     []Balance
	tokenList       []TokenData
	port            string = "8084"
	url             string = "http://localhost:8545"
	lastRefresh     time.Duration
	client          *ethclient.Client
	refreshDuration time.Duration = time.Second * 15
)

func init() {
	var err error
	portCustom := os.Getenv("PORT")
	duration := os.Getenv("DURATION")
	urlCustom := os.Getenv("GETH")
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
	if len(urlCustom) != 0 {
		fmt.Println("using custom geth:", urlCustom)
		url = urlCustom
	}
	if len(addresses) == 0 {
		panic("no addresses supplied")
	}
	client, err = ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	addressList = parseAddresses(addresses)
}

func init() {
	tokenlisturl := "https://raw.githubusercontent.com/Uniswap/default-token-list/main/src/tokens/mainnet.json"
	resp, err := http.Get(tokenlisturl)
	if err != nil {
		fmt.Printf("Could not get token list: %s\n", err)
	}
	tokenListData, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println("Err: error reading body of response, ", err)
	}
	json.Unmarshal(tokenListData, &tokenList)
	if err != nil {
		fmt.Println("Err: Decoding tokenlist, ", err)
	}
	for i := range tokenList {
		tokenList[i].realAddress = common.HexToAddress(tokenList[i].Address)
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
		resp = append(resp, fmt.Sprintf("crypto_balance{name=\"%s\",address=\"%s\",symbol=\"%s\"} %v", v.name, v.address, v.symbol, v.balance))
	}
	resp = append(resp, fmt.Sprintf("crypto_load_seconds %0.2f", lastRefresh.Seconds()))
	fmt.Fprintln(w, strings.Join(resp, "\n"))
}
