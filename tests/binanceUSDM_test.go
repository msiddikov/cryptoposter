package tests

import (
	"os"
	"testing"

	lvn "github.com/Lavina-Tech-LLC/lavinagopackage/v2"
	"github.com/Lavina-Tech-LLC/lavinagopackage/v2/conf"
	"github.com/joho/godotenv"
	"github.com/msiddikov/cryptoposter"
	"github.com/msiddikov/cryptoposter/exchanges/binance/binanceUsdm"
	"github.com/msiddikov/cryptoposter/types"
	"github.com/shopspring/decimal"
)

var (
	auth types.Auth
)

func init() {
	godotenv.Load(conf.GetPath() + "binance.env")
	auth = types.Auth{
		Key:     os.Getenv("KEY"),
		Secret:  os.Getenv("SECRET"),
		Testnet: lvn.Ternary(os.Getenv("TESTNET") == "TRUE", true, false),
	}
}

func TestAssets(t *testing.T) {

	res, err := binanceUsdm.GetAssets(&auth)
	if err != nil {
		t.Error(err)
	}
	lvn.Logger.Notice(res)
}

func TestNewOrder(t *testing.T) {
	res, err := binanceUsdm.NewOrder(auth, "BUY", "BTCUSDT", decimal.RequireFromString("0.001"))
	if err != nil {
		t.Error(err)
	}
	lvn.Logger.Notice(res)
}

func TestCancelOrder(t *testing.T) {
	res, err := binanceUsdm.CancelOrder(auth, 3281849405, "BTCUSDT")
	if err != nil {
		t.Error(err)
	}
	lvn.Logger.Notice(res)

}

func TestOrderInfo(t *testing.T) {
	res, err := binanceUsdm.OrderInfo(auth, 3281849405, "BTCUSDT")
	if err != nil {
		t.Error(err)
	}
	res.Parse()
	lvn.Logger.Notice(res)
}

func TestExecute(t *testing.T) {
	client, err := cryptoposter.New(cryptoposter.NewCLientOpts{
		Exchange: cryptoposter.BinanceUSDM,
		Key:      auth.Key,
		Secret:   auth.Secret,
		Testnet:  auth.Testnet,
	})
	if err != nil {
		t.Error(err)
	}

	res, err := client.ExecuteSync(types.ExecOptions{
		Side:     "SELL",
		Size:     0.161,
		Symbol:   "BTCUSDT",
		ResChan:  make(chan types.ExecRes, 10),
		Leverage: 50,
	})
	if err != nil {
		t.Error(err)
	}
	lvn.Logger.Notice(res)
}
