package types

import (
	"context"

	"github.com/frankrap/bybit-api/rest"
	"github.com/frankrap/deribit-api"
	ftx "github.com/go-numb/go-ftx/rest"
)

type (
	ExchangeAccount struct {
		Exchange   string
		Api        string
		Secret     string
		SubAccount string
		RepSymbol  string
	}

	Auth struct {
		Key        string
		Secret     string
		Subaccount string
		Testnet    bool
		FTX        *ftx.Client
		Bybit      *rest.ByBit
		Bitmex     *context.Context
		Deribit    *deribit.Client
	}
	ExecRes struct {
		Id       string
		Qty      float64
		AvgPrice float64
		OrderId  string
		Filled   bool
		Err      error
	}

	ExecOptions struct {
		Id       string
		Side     string
		Size     float64
		ResChan  chan ExecRes
		Symbol   string
		IsMarket bool
	}
	Quote struct {
		Bid float64
		Ask float64
	}
)
