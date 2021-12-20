package core

import (
	"context"
	"fmt"
	"math/big"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gelfand/contract-feed/internal/token"
)

// token method IDs
var (
	symbolMethodID = common.FromHex("0x95d89b41")
	nameMethodID   = common.FromHex("0x06fdde03")
)

var big10 = big.NewInt(10)

type Token struct {
	Address     common.Address
	Name        string
	Symbol      string
	TotalSupply *big.Int
}

func (t Token) ToMsg() string {
	// TODO: make function wrappers for bold text and for hyperlinks.
	return fmt.Sprintf("Address: <a href=\"https://etherscan.io/token/%s\"><b>%s</b></a>\n"+
		"Name: <b>%s</b>\n"+
		"Symbol: <b>%s</b>\n"+
		"TotalSupply: <b>%v</b>\n",
		t.Address.String(), t.Address.String(),
		t.Name,
		t.Symbol,
		t.TotalSupply,
	)
}

func (c *Coordinator) isToken(ctx context.Context, addr common.Address) bool {
	// two checks would be sufficiently enough.
	msg := ethereum.CallMsg{
		To:   &addr,
		Data: symbolMethodID,
	}
	if _, err := c.client.PendingCallContract(ctx, msg); err != nil {
		return false
	}

	msg = ethereum.CallMsg{
		To:   &addr,
		Data: nameMethodID,
	}

	if _, err := c.client.PendingCallContract(ctx, msg); err != nil {
		return false
	}

	return true
}

// getTokenData retrieves token data by it's address.
func (c *Coordinator) getTokenData(tokenAddr common.Address) (Token, error) {
	// I faced it already, but might try without it.
	t, err := token.NewTokenCaller(tokenAddr, c.client)
	if err != nil {
		return Token{}, err
	}

	name, err := t.Name(&bind.CallOpts{})
	if err != nil {
		return Token{}, fmt.Errorf("could not retrieve token name: %w", err)
	}
	if !utf8.ValidString(name) {
		name = ""
	}

	symbol, err := t.Symbol(&bind.CallOpts{})
	if err != nil {
		return Token{}, fmt.Errorf("could not retrieve token symbol: %w", err)
	}
	if !utf8.ValidString(symbol) {
		symbol = ""
	}

	totalSupply, err := t.TotalSupply(&bind.CallOpts{})
	if err != nil {
		return Token{}, fmt.Errorf("could not retrieve token total supply: %w", err)
	}

	// retrieve token decimals to ensure this is correct ERC20/ERC20-like token implementation.
	if _, err := t.Decimals(&bind.CallOpts{}); err != nil {
		return Token{}, fmt.Errorf("could not retrieve token decimals: %w", err)
	}
	return Token{
		Address:     tokenAddr,
		Name:        name,
		Symbol:      symbol,
		TotalSupply: totalSupply,
	}, nil
}
