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

type Address struct {
	name    string
	address common.Address
}

type Balance struct {
	Address
	balance      string
	tokenAddress string
	symbol       string
}

func walletLoop() {
	for range time.Tick(refreshDuration) {
		refreshBalances()
	}
}

func refreshBalances() {
	fmt.Print("Refreshing...")
	start := time.Now()
	for i, v := range balanceList {
		balanceList[i].balance = GetBalance(v.address).String()
	}
	lastRefresh = time.Since(start)
	fmt.Printf(" Completed (%s)", lastRefresh)
}

func getTokenBalance(tokenAddress, address common.Address) *big.Float {
	caller, err := NewTokenCaller(tokenAddress, eth)
	if err != nil {
		fmt.Println("Err on token address: ", tokenAddress)
		return nil
	}
	decimals, err := caller.Decimals(nil)
	if err != nil {
		fmt.Println("Err on token address: ", tokenAddress)
		return nil
	}
	balance, err := caller.BalanceOf(nil, address)
	if err != nil {
		fmt.Println("Err on token address: ", tokenAddress)
		return nil
	}
	return IntToDec(balance, float64(decimals))
}

func refreshAddressBalances() {
	newBalances := []Balance{}
	for _, v := range addressList {
		newBalances = append(newBalances, Balance{Address: v, symbol: "ETH"})

	}
	balanceList = newBalances
}

func GetBalance(address common.Address) *big.Float {
	balance, err := eth.BalanceAt(context.Background(), address, nil)
	if err != nil {
		fmt.Printf("Error fetching balance (%v)\n", address)
	}
	return WeiToEther(balance)
}

func IntToDec(u *big.Int, decimal float64) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(u), big.NewFloat(decimal))
}

func WeiToEther(wei *big.Int) *big.Float {
	return IntToDec(wei, params.Ether)
}

func parseAddresses(a string) []Address {
	addresses := []Address{}
	var name string
	var address common.Address
	var err error

	for _, v := range strings.Split(a, ",") {
		if common.IsHexAddress(v) {
			address = common.HexToAddress(v)
			name, err = ens.ReverseResolve(eth, address)
			if err == nil {
				fmt.Printf("Found ENS (%s) for address (%s)\n", name, address)
			} else {
				name = v
			}
		} else {
			name = v
			address, err := ens.Resolve(eth, v)
			//this might be weird cause many address potentially
			if err != nil {
				fmt.Printf("ERR: Address (%s) not a hex address or ENS domain\n", address)
				continue
			}
		}
		addresses = append(addresses, Address{name: name, address: address})
	}
	return addresses
}
