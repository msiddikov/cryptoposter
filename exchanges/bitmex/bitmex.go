package bitmex

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/msiddikov/cryptoposter/types"

	lvn "github.com/Lavina-Tech-LLC/lavinagopackage/v2"
	"github.com/antihax/optional"
	"github.com/qct/bitmex-go/swagger"
)

type (
	authWithClient struct {
		auth   context.Context
		client *swagger.APIClient
	}
)

func GetClient(auth *types.Auth) error {

	client := context.WithValue(context.TODO(), swagger.ContextAPIKey, swagger.APIKey{
		Key:    auth.Key,
		Secret: auth.Secret,
	})

	auth.Bitmex = &client
	_, err := GetBTCPrice(*auth)
	return err
}

func getApiClient(auth types.Auth) *swagger.APIClient {
	config := swagger.NewConfiguration()
	config.BasePath = lvn.Ternary(
		auth.Testnet,
		"https://testnet.bitmex.com/api/v1",
		"https://bitmex.com/api/v1",
	)
	return swagger.NewAPIClient(config)
}

func Execute(auth types.Auth, opts types.ExecOptions) {
	defer func() {
		close(opts.ResChan)
	}()

	apiClient := getApiClient(auth)
	authWC := authWithClient{
		auth:   *auth.Bitmex,
		client: apiClient,
	}

	size := float32(opts.Size)
	order, err := newOrder(authWC, opts.Side, opts.Symbol, size)
	if err != nil {
		log.Printf("%v", err)
		opts.ResChan <- types.ExecRes{Err: err}
		return
	}

	opts.ResChan <- types.ExecRes{
		Id:      opts.Id,
		OrderId: order.OrderID,
	}
	cycleCount := 0
	start := time.Now()
	for order.CumQty != size && order.OrdStatus != "Canceled" {
		cycleCount++

		order1, err := amendOrder(authWC, order)
		if err != nil {
			errS := changeType(err)
			if strings.Contains(string(errS.Body()), "Invalid ordStatus: Filled") {
				break
			}
			if strings.Contains(string(errS.Body()), "Invalid ordStatus: Canceled") {
				opts.ResChan <- types.ExecRes{Err: fmt.Errorf("Bitmex order has been cancelled manually")}
				return
			}
			log.Printf("%v", err)
			continue
		} else {
			order = order1
		}
	}

	if cycleCount != 0 {
		timeSpent := time.Now().Sub(start)
		fmt.Printf("Finished %v cycles in total %v ms, average is %v\n", cycleCount, timeSpent.Milliseconds(), timeSpent.Milliseconds()/int64(cycleCount))
	}

	orders, _, err := apiClient.OrderApi.OrderGetOrders(*auth.Bitmex, &swagger.OrderApiOrderGetOrdersOpts{Filter: optional.NewString("{\"orderID\": \"" + order.OrderID + "\"}")})
	if err != nil {
		log.Printf("%v", err)
		opts.ResChan <- types.ExecRes{Err: err}
		return
	}
	order = orders[0]

	opts.ResChan <- types.ExecRes{
		Id:       opts.Id,
		Qty:      float64(order.CumQty),
		AvgPrice: order.AvgPx,
		OrderId:  order.OrderID,
		Filled:   true,
	}
}

func amendOrder(auth authWithClient, order swagger.Order) (swagger.Order, error) {
	empty := swagger.Order{}
	book, res, err := auth.client.OrderBookApi.OrderBookGetL2(auth.auth, order.Symbol, &swagger.OrderBookApiOrderBookGetL2Opts{
		Depth: optional.NewFloat32(1),
	})
	checkLimit(*res)
	if err != nil {
		return empty, err
	}

	priceChanged := 0 // 0 not changed, 1 changed, edit with peg, 2 changed, edit with limit and peg
	k := lvn.Ternary(order.Side == "Buy", 1, 0)

	if book[k].Price != order.Price {
		priceChanged = 1
	} else if book[k].Size == order.LeavesQty {
		priceChanged = 2
	}

	if priceChanged == 0 {
		return order, nil
	}

	var leavesQty float32
	if priceChanged == 2 {
		orders, res, err := auth.client.OrderApi.OrderGetOrders(auth.auth, &swagger.OrderApiOrderGetOrdersOpts{Filter: optional.NewString("{\"orderID\": \"" + order.OrderID + "\"}")})
		checkLimit(*res)
		if err != nil {
			return empty, nil
		}
		leavesQty = orders[0].LeavesQty

		_, res, err = auth.client.OrderApi.OrderCancel(auth.auth, &swagger.OrderApiOrderCancelOpts{
			OrderID: optional.NewString(order.OrderID),
		})
		checkLimit(*res)
		if err != nil {
			return empty, err
		}

		order, errS := newOrder(auth, order.Side, order.Symbol, leavesQty)
		if errS != nil {
			return empty, errS
		}

		return order, nil
	}

	order, res, err = auth.client.OrderApi.OrderAmend(auth.auth, &swagger.OrderApiOrderAmendOpts{
		PegOffsetValue: optional.NewFloat64(0),
		OrderID:        optional.NewString(order.OrderID),
	})

	checkLimit(*res)
	if err != nil {
		return empty, err
	}

	return order, nil
}

func newOrder(auth authWithClient, side, symbol string, size float32) (swagger.Order, error) {
	side = strings.Title(strings.ToLower(side))

	param := swagger.OrderApiOrderNewOpts{
		Side:           optional.NewString(side),
		PegPriceType:   optional.NewString("PrimaryPeg"),
		ExecInst:       optional.NewString("Fixed"),
		OrderQty:       optional.NewFloat32(size),
		OrdType:        optional.NewString("Pegged"),
		TimeInForce:    optional.NewString("GoodTillCancel"),
		PegOffsetValue: optional.NewFloat64(0),
	}

	res, _, err := auth.client.OrderApi.OrderNew(auth.auth, symbol, &param)

	if err != nil {
		errS := changeType(err)
		log.Printf("%v", string(errS.Body()))
		return swagger.Order{}, fmt.Errorf(string(errS.Body()))
	}
	log.Printf("%v", res)
	return res, nil
}

func changeType(e interface{}) swagger.GenericSwaggerError {
	return e.(swagger.GenericSwaggerError)
}

func getPrice(auth authWithClient, symbol string) (types.Quote, error) {
	result := types.Quote{}

	res, _, err := auth.client.QuoteApi.QuoteGet(auth.auth, &swagger.QuoteApiQuoteGetOpts{
		Symbol:  optional.NewString(symbol),
		Reverse: optional.NewBool(true),
	})

	if err != nil {
		return result, err
	}
	if len(res) < 1 {
		return result, nil
	}
	result.Ask = res[0].AskPrice
	result.Bid = res[0].BidPrice
	return result, nil
}

func GetBTCPrice(auth types.Auth) (types.Quote, error) {
	return getPrice(authWithClient{
		auth:   *auth.Bitmex,
		client: getApiClient(auth),
	}, "XBTUSD")
}

func checkLimit(res http.Response) {
	if res.StatusCode == 429 {
		s := res.Header.Get("retry-after")[0]

		sInt, err := strconv.ParseInt(string(s), 0, 16)
		if err != nil {
			return
		}
		fmt.Printf("Cooling down bitmex for %v seconds\n", sInt)
		time.Sleep(time.Duration(sInt) * time.Second)
	}
}
