package binanceUsdm

import (
	lvn "github.com/Lavina-Tech-LLC/lavinagopackage/v2"
	"github.com/shopspring/decimal"
)

func roundSize(symbol string, size decimal.Decimal) decimal.Decimal {
	s := exInfo.GetSymbol(symbol)
	return size.Round(int32(s.QuantityPrecision))
}

func getNextPrice(symbol, side string, price decimal.Decimal) decimal.Decimal {
	s := exInfo.GetSymbol(symbol)
	step := lvn.Ternary(s.PricePrecision, 1/float64(s.PricePrecision), 0)
	stepDec := decimal.NewFromFloat(step)
	return lvn.Ternary(side == "BUY", price.Add(stepDec), price.Sub(stepDec))
}
