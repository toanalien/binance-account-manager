package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/spf13/viper"
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
		fmt.Println(asset)
	}

	res, err := client.NewGetAccountService().Do(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)

}
