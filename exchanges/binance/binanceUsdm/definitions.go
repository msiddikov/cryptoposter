package binanceUsdm

import (
	"github.com/msiddikov/cryptoposter/types"
	"github.com/shopspring/decimal"
)

type (
	accountInfo struct {
		Assets []asset
	}

	asset struct {
		Asset            string
		WalletBalance    string
		MarginBalance    string
		AvailableBalance string
	}

	order struct {
		OrderId           int
		AvgPrice          string
		AvgPriceParsed    decimal.Decimal
		Side              string // BUY/SELL
		Status            string
		Symbol            string
		CumQty            string
		CumQtyParsed      decimal.Decimal
		OrigQty           string
		OrigQtyParsed     decimal.Decimal
		ExecutedQty       string
		ExecutedQtyParsed decimal.Decimal
	}

	orderBook struct {
		Bids       [][]string
		Asks       [][]string
		BidsParsed [][]decimal.Decimal
		AsksParsed [][]decimal.Decimal
	}

	exchangeInfo struct {
		Symbols []symbolInfo
	}

	symbolInfo struct {
		Symbol            string
		ContractType      string
		PricePrecision    int
		QuantityPrecision int
	}
)

var (
	exchange = "BinanceUSDM"
)

func parseSlice(s [][]string) [][]decimal.Decimal {
	res := [][]decimal.Decimal{}
	for _, se := range s {
		el := []decimal.Decimal{}
		for _, see := range se {
			if see != "" {
				el = append(el, decimal.RequireFromString(see))
			}
		}
		res = append(res, el)
	}
	return res
}

func (o *orderBook) Parse() {
	o.AsksParsed = parseSlice(o.Asks)
	o.BidsParsed = parseSlice(o.Bids)
}

func (o *orderBook) GetQuote() types.Quote {
	o.Parse()
	ask := float64(0)
	bid := float64(0)

	if len(o.AsksParsed) != 0 && len(o.AsksParsed[0]) != 0 {
		ask, _ = o.AsksParsed[0][0].Float64()
	}
	if len(o.BidsParsed) != 0 && len(o.BidsParsed[0]) != 0 {
		bid, _ = o.BidsParsed[0][0].Float64()
	}

	return types.Quote{
		Ask: ask,
		Bid: bid,
	}
}

func (o *order) Parse() {
	if o.AvgPrice != "" {
		o.AvgPriceParsed = decimal.RequireFromString(o.AvgPrice)
	}
	if o.CumQty != "" {
		o.CumQtyParsed = decimal.RequireFromString(o.CumQty)
	}
	if o.OrigQty != "" {
		o.OrigQtyParsed = decimal.RequireFromString(o.OrigQty)
	}
	if o.ExecutedQty != "" {
		o.ExecutedQtyParsed = decimal.RequireFromString(o.ExecutedQty)
	}
}

func (i *exchangeInfo) GetSymbol(symbol string) symbolInfo {
	for _, s := range i.Symbols {
		if s.Symbol == symbol {
			return s
		}
	}

	return symbolInfo{}
}
