package exchanges

import "cryptoposter/exchanges/definitions"

type (
	integration struct {
		acc       definitions.ExchangeAccount
		auth      Auth
		getClient func(key, secret, subAccount string) (Auth, error)
		Execute   func(Auth)
	}
)
