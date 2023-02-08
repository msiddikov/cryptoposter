package cryptoposter

import (
	"fmt"

	"github.com/msiddikov/cryptoposter/exchanges/binance/binanceUsdm"
	"github.com/msiddikov/cryptoposter/exchanges/bitmex"
	"github.com/msiddikov/cryptoposter/exchanges/deribit"
	"github.com/msiddikov/cryptoposter/types"
)

func New(o NewCLientOpts) (CryptoPoster, error) {
	res := CryptoPoster{}

	res.config = config{
		auth: types.Auth{
			Key:        o.Key,
			Secret:     o.Secret,
			Subaccount: o.Subaccount,
			Testnet:    o.Testnet,
		},
		exchange: o.Exchange,
	}
	var err error

	switch o.Exchange {
	case Bitmex:
		err = bitmex.GetClient(&res.config.auth)
		res.execute = bitmex.Execute
	case Deribit:
		err = deribit.GetClient(&res.config.auth)
		res.execute = deribit.Execute
	case BinanceUSDM:
		err = binanceUsdm.Test(res.config.auth)
		res.execute = binanceUsdm.Execute
	default:
		return CryptoPoster{}, fmt.Errorf("no such exchange")
	}

	if err != nil {
		return CryptoPoster{}, nil
	}

	return res, nil
}

func (cp *CryptoPoster) Execute(opts types.ExecOptions) {
	cp.execute(cp.config.auth, opts)
}

func (cp *CryptoPoster) ExecuteSync(opts types.ExecOptions) (types.ExecRes, error) {
	cp.execute(cp.config.auth, opts)

	for res := range opts.ResChan {
		if res.Filled || res.Err != nil {
			return res, res.Err
		}
	}
	return types.ExecRes{}, fmt.Errorf("channel unexpectedly closed")
}

func GetAvailableExchanges() []Exchange {
	return []Exchange{
		Deribit,
		Bitmex,
		BinanceUSDM,
	}
}
