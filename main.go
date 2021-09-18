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
	client := binance.NewClient(apiKey, secretKey)
	res, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res.Balances)
}
