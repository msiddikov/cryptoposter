package types

type (
	ExecRes struct {
		Id       string
		Qty      float64
		AvgPrice float64
		OrderId  string
		Filled   bool
		Err      error
	}

	ExecOptions struct {
		Id        string
		Side      string
		Size      float64
		ResChan   chan ExecRes
		StartChan chan bool
		Market    Market
	}

	Market struct {
		Exchange string
		Symbol   string
		IsMarket bool
	}
)
