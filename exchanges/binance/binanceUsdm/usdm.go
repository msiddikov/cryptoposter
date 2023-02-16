package binanceUsdm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	lvn "github.com/Lavina-Tech-LLC/lavinagopackage/v2"
	"github.com/msiddikov/cryptoposter/types"
	"github.com/shopspring/decimal"
)

func Test(auth types.Auth) error {
	_, err := GetAssets(&auth)
	return err
}

func Execute(auth types.Auth, opts types.ExecOptions) {
	defer func() {
		close(opts.ResChan)
	}()

	err := SetLeverage(auth, opts.Symbol, opts.Leverage)
	if err != nil {
		lvn.Logger.Errorf("%v", err)
		opts.ResChan <- types.ExecRes{Err: err}
		return
	}

	order := order{}
	sum := decimal.NewFromInt(0)
	qty := decimal.NewFromInt(0)

	size := roundSize(opts.Symbol, decimal.NewFromFloat(opts.Size))
	remQty := size

	cycleCount := 0
	start := time.Now()
	for remQty.GreaterThan(decimal.Zero) {
		order, err = NewOrder(auth, opts.Side, opts.Symbol, remQty)
		if err != nil {
			lvn.Logger.Errorf("%v", err)
			opts.ResChan <- types.ExecRes{Err: err}
			return
		}
		CancelOrder(auth, order.OrderId, opts.Symbol)
		order, err = OrderInfo(auth, order.OrderId, opts.Symbol)
		if err != nil {
			lvn.Logger.Errorf("%v", err)
			opts.ResChan <- types.ExecRes{Err: err}
			return
		}
		order.Parse()
		qty = qty.Add(order.ExecutedQtyParsed)
		sumCur := order.ExecutedQtyParsed.Mul(order.AvgPriceParsed)
		sum = sum.Add(sumCur)
		remQty = roundSize(opts.Symbol, remQty.Sub(order.ExecutedQtyParsed))
		cycleCount++
	}
	if cycleCount != 0 {
		timeSpent := time.Since(start)
		fmt.Printf("Finished %v cycles in total %v ms, average is %v\n", cycleCount, timeSpent.Milliseconds(), timeSpent.Milliseconds()/int64(cycleCount))
	}

	avgPrice := sum.Div(qty)
	opts.ResChan <- types.ExecRes{
		Id:       opts.Id,
		Qty:      qty.InexactFloat64(),
		AvgPrice: avgPrice.InexactFloat64(),
		OrderId:  fmt.Sprint(order.OrderId),
		Filled:   true,
	}
}

func GetAssets(auth *types.Auth) ([]types.Asset, error) {
	res := []types.Asset{}
	resp, err := fetch[accountInfo](auth, reqParams{
		Method:   "GET",
		Endpoint: "/fapi/v2/account",
	})
	if err != nil {
		return []types.Asset{}, err
	}
	for _, a := range resp.Resp.Assets {
		amount, err := strconv.ParseFloat(a.AvailableBalance, 64)
		if err != nil {
			return res, err
		}
		if amount == 0 {
			continue
		}
		res = append(res, types.Asset{
			Symbol:   a.Asset,
			Amount:   amount,
			Exchange: exchange,
		})
	}
	return res, nil
}

func NewOrder(auth types.Auth, side, symbol string, size decimal.Decimal) (order, error) {
	orderBook, err := GetPrice(auth, symbol)
	if err != nil {
		return order{}, err
	}

	size = roundSize(symbol, size)
	if size.String() == "0" {
		return order{}, fmt.Errorf("order size cannot be 0")
	}

	price := lvn.Ternary(side == "BUY", orderBook.Bid, orderBook.Ask)
	priceDec := getNextPrice(symbol, side, decimal.NewFromFloat(price))

	resp, err := fetch[order](&auth, reqParams{
		Method:     "POST",
		Endpoint:   "/fapi/v1/order",
		ForceQuery: true,
		QParams: []queryParam{
			{
				Key:   "type",
				Value: "LIMIT",
			},
			{
				Key:   "symbol",
				Value: symbol,
			},
			{
				Key:   "side",
				Value: side,
			},
			{
				Key:   "quantity",
				Value: size.String(),
			},
			{
				Key:   "price",
				Value: priceDec.String(),
			},
			{
				Key:   "timeInForce",
				Value: "GTX",
			},
		},
	})

	if err != nil && strings.Contains(err.Error(), "the Post Only order will be rejected") {
		return NewOrder(auth, side, symbol, size)
	}
	return resp.Resp, err

}

func GetPrice(auth types.Auth, symbol string) (types.Quote, error) {
	resp, err := fetch[orderBook](&auth, reqParams{
		Method:   "GET",
		Endpoint: "/fapi/v1/depth",
		QParams: []queryParam{
			{
				Key:   "symbol",
				Value: symbol,
			},
			{
				Key:   "limit",
				Value: "5",
			},
		},
	})
	return resp.Resp.GetQuote(), err
}
func SetLeverage(auth types.Auth, symbol string, lvrg int) error {
	_, err := fetch[struct{}](&auth, reqParams{
		Method:     "POST",
		Endpoint:   "/fapi/v1/leverage",
		ForceQuery: true,
		QParams: []queryParam{
			{
				Key:   "symbol",
				Value: symbol,
			},
			{
				Key:   "leverage",
				Value: fmt.Sprint(lvrg),
			},
		},
	})
	return err
}

func CancelOrder(auth types.Auth, orderId int, symbol string) (order, error) {
	resp, err := fetch[order](&auth, reqParams{
		Method:     "DELETE",
		Endpoint:   "/fapi/v1/order",
		ForceQuery: true,
		QParams: []queryParam{
			{
				Key:   "symbol",
				Value: symbol,
			},
			{
				Key:   "orderId",
				Value: fmt.Sprint(orderId),
			},
		},
	})

	return resp.Resp, err
}

func OrderInfo(auth types.Auth, orderId int, symbol string) (order, error) {
	resp, err := fetch[order](&auth, reqParams{
		Method:     "GET",
		Endpoint:   "/fapi/v1/order",
		ForceQuery: true,
		QParams: []queryParam{
			{
				Key:   "symbol",
				Value: symbol,
			},
			{
				Key:   "orderId",
				Value: fmt.Sprint(orderId),
			},
		},
	})

	return resp.Resp, err
}
