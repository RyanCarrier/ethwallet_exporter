package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	log "github.com/sirupsen/logrus"
	ens "github.com/wealdtech/go-ens/v3"
)

// TokenData keeps track of information one a specific token
type TokenData struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	realAddress common.Address
	Symbol      string `json:"symbol"`
	Decimals    uint8  `json:"decimals"`
	ChainID     uint8  `json:"chainId"`
	LogoURI     string `json:"logoURI"`
}

// Address holds information of token balances of a wallet(address)
type Address struct {
	name     string
	address  common.Address
	balances []Balance
}

// Balance specifies the balance/amount of a token
type Balance struct {
	balance string
	token   TokenData
	symbol  string
}

// walletLoop runs every tick, scanning all tokens to check for every cacheTicks, and refreshes known balances every tick
func walletLoop() {
	var i uint = 0
	refreshAllTokens()
	for range time.Tick(refreshDuration) {
		if i >= cacheTicks {
			refreshAllTokens()
			i = 0
		} else {
			refreshKnownBalances()
			i++
		}
	}
}

func refreshKnownBalances() {
	start := time.Now()
	total := 0
	for i, v := range addressList {
		for j, jv := range v.balances {
			if (jv.token == TokenData{}) {
				addressList[i].balances[j].balance = getEthBalance(v.address).String()
			} else {
				addressList[i].balances[j].balance = getTokenBalance(jv.token, v.address).String()
			}
		}
		total += len(v.balances)
	}
	lastRefresh = time.Since(start)
	log.Infof("Refreshed %d addresses (%d balances) (%s)", len(addressList), total, lastRefresh)
}

// RefreshAllTokens checks all available tokens for non-zero balances
func refreshAllTokens() {
	start := time.Now()
	if len(addressList) < len(rawAddresses) {
		log.Warn("Address list doesn't appear fully loaded, redailing geth")
		connectClient()
		addressList = parseAddresses(rawAddresses)
	}
	for i, v := range addressList {
		addressList[i].balances = []Balance{{symbol: "ETH", balance: getEthBalance(v.address).String()}}
		for _, jv := range tokenList {
			bal := getTokenBalance(jv, v.address)
			if bal.Cmp(big.NewFloat(0)) != 0 {
				addressList[i].balances = append(addressList[i].balances, Balance{token: jv, symbol: jv.Symbol, balance: bal.String()})
			}
		}
	}
	lastRefresh = time.Since(start)
	log.Infof("Refreshed %d addresses and scanned for %d tokens (%s)", len(addressList), len(tokenList), lastRefresh)
}

func getEthBalance(address common.Address) *big.Float {
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Errorf("Error fetching balance (%v)", address)
		log.Info("Attempting to redail to geth...")
		connectClient()
		addressList = parseAddresses(rawAddresses)
	}
	return weiToEther(balance)
}

func getTokenBalance(token TokenData, address common.Address) *big.Float {
	caller, err := NewTokenCaller(token.realAddress, client)
	if err != nil {
		fmt.Println("Err on token address: ", token.realAddress)
		return big.NewFloat(0)
	}
	balance, err := caller.BalanceOf(nil, address)
	if err != nil {
		log.Error("Err on token address: ", token.realAddress)
		return big.NewFloat(0)
	}
	return intToDec(balance, token.Decimals)
}

func intToDec(u *big.Int, decimal uint8) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(u),
		new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)))
}

func weiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

// parseAddresses converts address strings into Address structs, so we can handle hex wallets and ENS domains
func parseAddresses(addressSlice []string) []Address {
	addresses := []Address{}
	var name string
	var address common.Address
	var err error

	for _, v := range addressSlice {
		if common.IsHexAddress(v) {
			address = common.HexToAddress(v)
			name, err = ens.ReverseResolve(client, address)
			if err == nil {
				log.Infof("Found ENS (%s) for address (%s)", name, address)
			} else {
				name = v
			}
		} else {
			log.Infof("'%s' does not appear to be hex address attempting to resolve...", v)
			name = v
			address, err = ens.Resolve(client, v)
			//this might be weird cause many address potentially? for doge btc etc
			if err != nil {
				log.Error("ERROR: getting from ENS", err.Error())
				log.Errorf("ERR: Address (%s) not a hex address or ENS domain", v)
				continue
			}
			log.Infof("Name (%s) successfully resolved to address (%s)", v, address)
		}
		addresses = append(addresses, Address{name: name, address: address})
	}
	return addresses
}
