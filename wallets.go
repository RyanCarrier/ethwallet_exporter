package main

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	ens "github.com/wealdtech/go-ens/v3"
)

type TokenData struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	realAddress common.Address
	Symbol      string `json:"symbol"`
	Decimals    uint8  `json:"decimals"`
	ChainId     uint8  `json:"chainId"`
	LogoURI     string `json:"logoURI"`
}
type Address struct {
	name     string
	address  common.Address
	balances []Balance
}

type Balance struct {
	balance string
	token   TokenData
	symbol  string
}

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
	fmt.Printf("Refreshed %d addresses (%d balances) (%s)\n", len(addressList), total, lastRefresh)
}

func refreshAllTokens() {
	start := time.Now()
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
	fmt.Printf("Refreshed %d addresses and scanned for %d tokens (%s)\n", len(addressList), len(tokenList), lastRefresh)
}

func getEthBalance(address common.Address) *big.Float {
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		fmt.Printf("Error fetching balance (%v)\n", address)
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
		fmt.Println("Err on token address: ", token.realAddress)
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
				fmt.Printf("Found ENS (%s) for address (%s)\n", name, address)
			} else {
				name = v
			}
		} else {
			fmt.Printf("'%s' does not appear to be hex address attempting to resolve...\n", v)
			name = v
			address, err = ens.Resolve(client, v)
			//this might be weird cause many address potentially? for doge btc etc
			if err != nil {
				fmt.Println("ERROR: getting from ENS", err.Error())
				fmt.Printf("ERR: Address (%s) not a hex address or ENS domain\n", v)
				continue
			}
			fmt.Printf("Name (%s) successfully resolved to address (%s)\n", v, address)
		}
		addresses = append(addresses, Address{name: name, address: address})
	}
	return addresses
}
