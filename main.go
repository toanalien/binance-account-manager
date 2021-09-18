package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/spf13/viper"
	"strconv"
)

func main() {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()
	var (
		apiKey    = viper.Get("API_KEY").(string)
		secretKey = viper.Get("SECRET_KEY").(string)
	)
	ctx := context.Background()

	client := binance.NewClient(apiKey, secretKey)
	margin, _ := client.NewGetIsolatedMarginAccountService().Do(ctx)
	userAssets := margin.Assets

	for _, asset := range userAssets {
		liquidatePrice, err := strconv.ParseFloat(asset.LiquidatePrice, 64)
		if (err == nil) && (liquidatePrice != 0) {
			fmt.Printf("%s %f\n", asset.Symbol, liquidatePrice)
		}

	}
}
