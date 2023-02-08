package cryptoposter

import "github.com/msiddikov/cryptoposter/types"

type (
	CryptoPoster struct {
		config  config
		opts    NewCLientOpts
		execute func(types.Auth, types.ExecOptions)
	}

	config struct {
		exchange exchange
		auth     types.Auth
	}

	NewCLientOpts struct {
		Exchange   exchange
		Key        string
		Secret     string
		Subaccount string
		Testnet    bool
	}

	exchange int
)

const (
	Deribit exchange = iota
	Bitmex
	BinanceUSDM
)
