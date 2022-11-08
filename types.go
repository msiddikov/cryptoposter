package cryptoposter

type (
	CryptoPoster struct {
		config Config
		opts   NewCLientOpts
	}

	Config struct {
		Exchange exchange
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
