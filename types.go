package cryptoposter

import "github.com/msiddikov/cryptoposter/types"

type (
	CryptoPoster struct {
		config  config
		opts    NewCLientOpts
		execute func(types.Auth, types.ExecOptions)
	}

	config struct {
		exchange Exchange
		auth     types.Auth
	}

	NewCLientOpts struct {
		Exchange   Exchange
		Key        string
		Secret     string
		Subaccount string
		Testnet    bool
	}

	Exchange int
)

const (
	Deribit Exchange = iota
	Bitmex
	BinanceUSDM
)
