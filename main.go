package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	rawAddresses    []string
	addressList     []Address = make([]Address, 0)
	tokenList       []TokenData
	port            int
	url             string
	lastRefresh     time.Duration
	client          *ethclient.Client
	refreshDuration time.Duration = time.Second * 15
	cacheTicks      uint
)

func init() {
	flag.IntVar(&port, "port", 9887, "Port to listen for http requests")
	flag.DurationVar(&refreshDuration, "duration", time.Second*15, "Duration between re-scanning for balance changes")
	flag.StringVar(&url, "geth", "http://localhost:8545", "Path to geth RPC")
	flag.StringSliceVar(&rawAddresses, "addresses", []string{"vitalik.eth", "0xEA674fdDe714fd979de3EdF0F56AA9716B898ec8"}, "\"address1.eth,0xDEADBEEF\"")
	flag.UintVar(&cacheTicks, "cache", 4, "Sets amount of balance refreshes (of previously known balances) before re-scanning all potential tokens, set to 0 to always scan every token (slower)")
	flag.Parse()
	if len(rawAddresses) == 0 {
		log.Panic("no addresses supplied")
	}
	importTokenList()
	connectClient()
	addressList = parseAddresses(rawAddresses)
}

func main() {
	go walletLoop()
	http.HandleFunc("/metrics", handleMetrics)
	panic(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), nil))
}

func connectClient() {
	var err error
	client, err = ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
}

// importTokenList pulls down known possible tokens (uses uniswaps, there could be more but this is pretty much all)
func importTokenList() {
	tokenlisturl := "https://raw.githubusercontent.com/Uniswap/default-token-list/main/src/tokens/mainnet.json"
	resp, err := http.Get(tokenlisturl)
	if err != nil {
		log.Errorf("Could not get token list: %s", err)
	}
	tokenListData, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Error("Err: error reading body of response", err)
	}
	json.Unmarshal(tokenListData, &tokenList)
	if err != nil {
		log.Error("Err: Decoding tokenlist", err)
	}
	for i := range tokenList {
		tokenList[i].realAddress = common.HexToAddress(tokenList[i].Address)
	}
}

// handleMetrics is for the prometheus exporter, handling their requests
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	var resp []string
	for _, v := range addressList {
		for _, b := range v.balances {
			if b.balance == "" {
				b.balance = "0"
			}
			resp = append(resp, fmt.Sprintf("crypto_balance{name=\"%s\",address=\"%s\",symbol=\"%s\"} %v", v.name, v.address, b.symbol, b.balance))
		}
	}
	resp = append(resp, fmt.Sprintf("crypto_load_seconds %0.2f", lastRefresh.Seconds()))
	fmt.Fprintln(w, strings.Join(resp, "\n"))
}
