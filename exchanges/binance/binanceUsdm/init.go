package binanceUsdm

import "github.com/msiddikov/cryptoposter/types"

var (
	exInfo exchangeInfo
)

func init() {
	resp, err := fetch[exchangeInfo](&types.Auth{}, reqParams{
		Method:   "GET",
		Endpoint: "/fapi/v1/exchangeInfo",
	})
	if err != nil {
		panic("Unable to get exchange info from binance futures")
	}
	exInfo = resp.Resp
}
