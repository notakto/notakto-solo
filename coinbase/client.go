package coinbase

import (
	"net/http"
	"time"

	commerce "github.com/coinbase-samples/commerce-sdk-go"
)

func NewCommerceClient(apiKey string) *commerce.Client {
	creds := &commerce.Credentials{ApiKey: apiKey}
	httpClient := http.Client{Timeout: 10 * time.Second}
	return commerce.NewClient(creds, httpClient)
}
