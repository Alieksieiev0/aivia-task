package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aiviaio/go-binance/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	var (
		apiKey        = "your api key"
		secretKey     = "your secret key"
		g             = new(errgroup.Group)
		symbolsPrices = make(chan map[string]string)
	)
	client := binance.NewClient(apiKey, secretKey)
	symbols, err := getFirstFiveSymbols(client)
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range symbols {
		g.Go(func() error {
			price, err := getSymbolPrice(client, s)
			if err != nil {
				return err
			}
			symbolsPrices <- price
			return nil
		})
	}

	go func() {
		if err := g.Wait(); err != nil {
			fmt.Printf("Error reading prices %v", err)
		}
		close(symbolsPrices)
	}()

	for symbolPrice := range symbolsPrices {
		for k, v := range symbolPrice {
			fmt.Println(k, v)
		}
	}
}

func getFirstFiveSymbols(client *binance.Client) ([]binance.Symbol, error) {
	service := client.NewExchangeInfoService()
	res, err := service.Do(context.Background())
	if err != nil {
		return nil, err
	}

	return res.Symbols[:5], nil
}

func getSymbolPrice(client *binance.Client, symbol binance.Symbol) (map[string]string, error) {
	service := client.NewListPricesService()
	service.Symbol(symbol.Symbol)
	res, err := service.Do(context.Background())
	if err != nil {
		return nil, err
	}

	if len(res) == 0 || res[0] == nil {
		return nil, fmt.Errorf("symbol price was not fetched from API correctly")
	}

	price := res[0]
	return map[string]string{price.Symbol: price.Price}, nil
}
