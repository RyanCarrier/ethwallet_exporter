package main

import (
	"context"
	"fmt"
	"math/big"
	"strings"
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
	name    string
	address common.Address
}

type Balance struct {
	Address
	balance string
	token   TokenData
	symbol  string
}

func walletLoop() {
	for range time.Tick(refreshDuration) {
		refreshBalances()
	}
}

func refreshBalances() {
	fmt.Printf("Refreshing %d addresses and %d balances... ", len(addressList), len(balanceList))
	start := time.Now()
	refreshAddressBalances()
	for i, v := range balanceList {
		if (v.token == TokenData{}) {
			balanceList[i].balance = GetBalance(v.address).String()
		} else {
			balanceList[i].balance = getTokenBalance(v.token, v.address).String()
		}

	}
	lastRefresh = time.Since(start)
	fmt.Printf(" Completed (%s)\n", lastRefresh)
}

func getTokenBalance(token TokenData, address common.Address) *big.Float {
	caller, err := NewTokenCaller(token.realAddress, client)
	if err != nil {
		fmt.Println("Err on token address: ", token.realAddress)
		return nil
	}

	balance, err := caller.BalanceOf(nil, address)
	if err != nil {
		fmt.Println("Err on token address: ", token.realAddress)
		return nil
	}
	return IntToDec(balance, token.Decimals)
}

func refreshAddressBalances() {
	newBalances := []Balance{}
	for _, v := range addressList {
		newBalances = append(newBalances, Balance{Address: v, symbol: "ETH"})
		for _, jv := range tokenList {
			bal := getTokenBalance(jv, v.address)
			if bal.Cmp(big.NewFloat(0)) != 0 {
				newBalances = append(newBalances, Balance{Address: v, token: jv, symbol: jv.Symbol})
			}
		}
	}
	balanceList = newBalances
}

func GetBalance(address common.Address) *big.Float {
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		fmt.Printf("Error fetching balance (%v)\n", address)
	}
	return WeiToEther(balance)
}

func IntToDec(u *big.Int, decimal uint8) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(u),
		new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimal)), nil)))
}

func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

func parseAddresses(a string) []Address {
	addresses := []Address{}
	var name string
	var address common.Address
	var err error

	for _, v := range strings.Split(a, ",") {
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
