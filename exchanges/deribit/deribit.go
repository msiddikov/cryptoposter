package deribit

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/msiddikov/cryptoposter/types"

	lvn "github.com/Lavina-Tech-LLC/lavinagopackage/v2"
	"github.com/frankrap/deribit-api"
	"github.com/frankrap/deribit-api/models"
)

type (
	getInstrumentParam struct {
		InstrumentName string `json:"instrument_name"`
	}
)

func GetClient(auth *types.Auth) error {
	cfg := &deribit.Configuration{
		ApiKey:        auth.Key,
		SecretKey:     auth.Secret,
		AutoReconnect: true,
		DebugMode:     false,
	}
	cfg.Addr = lvn.Ternary(
		auth.Testnet,
		deribit.TestBaseURL,
		deribit.RealBaseURL,
	)

	auth.Deribit = deribit.New(cfg)

	_, err := auth.Deribit.Test()
	return err
}

func Execute(auth types.Auth, opts types.ExecOptions) {
	defer func() {
		close(opts.ResChan)
	}()

	order, err := newOrder(auth, opts.Side, opts.Symbol, opts.Size)
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
	for order.FilledAmount != opts.Size {
		cycleCount++
		editedOrder, err := editOrder(auth, opts.Side, opts.Symbol, order.OrderID, opts.Size)
		if err != nil {
			if err.Error() == "jsonrpc2: code 11044 message: not_open_order" {
				break
			}

			if err.Error() == "jsonrpc2: code 10028 message: too_many_requests" {
				time.Sleep(1 * time.Second)
				continue
			}

			log.Printf("%v", err.Error())
			continue
		}
		order = editedOrder
	}

	if cycleCount != 0 {
		timeSpent := time.Now().Sub(start)
		fmt.Printf("Finished %v cycles in total %v ms, average is %v\n", cycleCount, timeSpent.Milliseconds(), timeSpent.Milliseconds()/int64(cycleCount))
	}

	order, err = auth.Deribit.GetOrderState(&models.GetOrderStateParams{OrderID: order.OrderID})
	if err != nil {
		log.Printf("%v", err)
		opts.ResChan <- types.ExecRes{Err: err}
		return
	}

	opts.ResChan <- types.ExecRes{
		Id:       opts.Id,
		Qty:      order.FilledAmount,
		AvgPrice: order.AveragePrice,
		OrderId:  order.OrderID,
		Filled:   true,
	}
}

func newOrder(auth types.Auth, side, symbol string, size float64) (models.Order, error) {
	if side == "BUY" {
		resp, err := buyOrder(auth, symbol, size)
		if err != nil {
			log.Printf("%v", err)
			return models.Order{}, err
		}
		return resp.Order, nil
	} else if side == "SELL" {
		resp, err := sellOrder(auth, symbol, size)
		if err != nil {
			log.Printf("%v", err)
			return models.Order{}, err
		}
		return resp.Order, nil
	}
	return models.Order{}, fmt.Errorf("Invalid side notation")
}

func editOrder(auth types.Auth, side, symbol, orderId string, size float64) (models.Order, error) {
	tickParams := &models.TickerParams{
		InstrumentName: symbol,
	}
	price, err := auth.Deribit.Ticker(tickParams)
	if err != nil {
		log.Printf("%v", err)
		return models.Order{}, err
	}

	instrument, err := getInstrument(auth, symbol)
	if err != nil {
		log.Printf("%v", err)
		return models.Order{}, err
	}

	var offset float64
	if side == "BUY" {
		offset = 1.05
	} else if side == "SELL" {
		offset = 0.95
	}

	param := models.EditParams{
		OrderID:  orderId,
		Price:    math.Round(price.IndexPrice*offset/instrument.TickSize) * instrument.TickSize,
		PostOnly: true,
		Amount:   size,
	}
	res, err := auth.Deribit.Edit(&param)
	if err != nil {
		log.Printf("%v", err)
		return models.Order{}, err
	}
	return res.Order, nil

}

func buyOrder(auth types.Auth, symbol string, size float64) (models.BuyResponse, error) {
	tickParams := &models.TickerParams{
		InstrumentName: symbol,
	}
	price, err := auth.Deribit.Ticker(tickParams)
	if err != nil {
		log.Printf("%v", err)
		return models.BuyResponse{}, err
	}

	instrument, err := getInstrument(auth, symbol)
	if err != nil {
		log.Printf("%v", err)
		return models.BuyResponse{}, err
	}

	buyParams := &models.BuyParams{
		InstrumentName: symbol,
		Amount:         size,
		Price:          math.Round(price.IndexPrice*1.05/instrument.TickSize) * instrument.TickSize,
		Type:           "limit",
		PostOnly:       true,
	}
	buyResult, err := auth.Deribit.Buy(buyParams)
	if err != nil {
		log.Printf("%v", err)
		return models.BuyResponse{}, err
	}
	log.Printf("%v", buyResult)
	return buyResult, nil
}

func sellOrder(auth types.Auth, symbol string, size float64) (models.SellResponse, error) {
	tickParams := &models.TickerParams{
		InstrumentName: symbol,
	}
	price, err := auth.Deribit.Ticker(tickParams)
	if err != nil {
		log.Printf("%v", err)
		return models.SellResponse{}, err
	}

	instrument, err := getInstrument(auth, symbol)
	if err != nil {
		log.Printf("%v", err)
		return models.SellResponse{}, err
	}

	sellParams := &models.SellParams{
		InstrumentName: symbol,
		Amount:         size,
		Price:          math.Round(price.IndexPrice*0.95/instrument.TickSize) * instrument.TickSize,
		Type:           "limit",
		PostOnly:       true,
	}
	sellResult, err := auth.Deribit.Sell(sellParams)
	if err != nil {
		log.Printf("%v", err)
		return models.SellResponse{}, err
	}
	log.Printf("%v", sellResult)
	return sellResult, nil
}

func getInstrument(auth types.Auth, symbol string) (result models.Instrument, err error) {
	params := getInstrumentParam{InstrumentName: symbol}
	err = auth.Deribit.Call("public/get_instrument", params, &result)
	return
}

func getPrice(auth types.Auth, symbol string) (types.Quote, error) {

	params := models.GetOrderBookParams{
		InstrumentName: symbol,
	}
	res, err := auth.Deribit.GetOrderBook(&params)
	if err != nil {
		return types.Quote{}, err
	}

	return types.Quote{
		Ask: res.BestAskPrice,
		Bid: res.BestBidPrice,
	}, nil
}
