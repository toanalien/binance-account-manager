package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"os"
	"sort"
	"strconv"
	"time"
)

func main() {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()
	var (
		apiKey    = viper.Get("API_KEY").(string)
		secretKey = viper.Get("SECRET_KEY").(string)
	)
	ctx := context.Background()

	binanceClient := binance.NewClient(apiKey, secretKey)
	isolatedMargin, _ := binanceClient.NewGetIsolatedMarginAccountService().Do(ctx)
	userAssets := isolatedMargin.Assets

	sort.Slice(userAssets, func(i, j int) bool {
		baseAssetNetAssetOfBtc, _ := strconv.ParseFloat(userAssets[i].BaseAsset.NetAssetOfBtc, 64)
		quoteAssetNetAssetOfBtc, _ := strconv.ParseFloat(userAssets[i].QuoteAsset.NetAssetOfBtc, 64)
		netAssetOfBtcI := baseAssetNetAssetOfBtc + quoteAssetNetAssetOfBtc
		baseAssetNetAssetOfBtc, _ = strconv.ParseFloat(userAssets[j].BaseAsset.NetAssetOfBtc, 64)
		quoteAssetNetAssetOfBtc, _ = strconv.ParseFloat(userAssets[j].QuoteAsset.NetAssetOfBtc, 64)
		netAssetOfBtcJ := baseAssetNetAssetOfBtc + quoteAssetNetAssetOfBtc
		return netAssetOfBtcI > netAssetOfBtcJ
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Symbol", "QuoteAsset", "BaseAsset", "LiquidatePrice"})
	for _, v := range userAssets {
		liquidatePrice, _ := strconv.ParseFloat(v.LiquidatePrice, 64)
		quoteAsset, _ := strconv.ParseFloat(v.QuoteAsset.TotalAsset, 64)
		baseAsset, _ := strconv.ParseFloat(v.BaseAsset.TotalAsset, 64)
		if liquidatePrice != 0 || quoteAsset != 0 || baseAsset != 0 {
			table.Append([]string{v.Symbol, v.QuoteAsset.NetAsset, v.BaseAsset.NetAsset, v.LiquidatePrice})
		}
	}
	table.Render()

	isolatedMarginOpenOrders, err := binanceClient.NewListMarginOpenOrdersService().IsIsolated(true).Symbol("BTCUSDT").Do(ctx)

	if err != nil {

	}

	sort.Slice(isolatedMarginOpenOrders, func(i, j int) bool {
		return isolatedMarginOpenOrders[i].Time > isolatedMarginOpenOrders[j].Time
	})

	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Symbol", "ExecutedQuantity", "QuoteQuantity", "Side", "Status", "Type", "TimeInForce"})
	for _, v := range isolatedMarginOpenOrders {
		table.Append([]string{v.Symbol, v.ExecutedQuantity, v.CummulativeQuoteQuantity, string(v.Side), string(v.Status), string(v.Type), time.UnixMilli(v.Time).Format("2006-01-02 15:04:05 -0700")})
	}
	table.Render()

	isolatedMarginTrades, err := binanceClient.NewListMarginTradesService().IsIsolated(true).Symbol("BTCUSDT").Do(ctx)
	if err != nil {

	}

	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Symbol", "IsBuyer", "Price", "CommissionAsset", "Commission", "Quantity", "QuoteQuantity", "Time"})

	sort.Slice(isolatedMarginTrades, func(i, j int) bool {
		return isolatedMarginTrades[i].Time > isolatedMarginTrades[j].Time
	})

	for _, v := range isolatedMarginTrades {
		side := "SELL"
		if v.IsBuyer == true {
			side = "BUY"
		}
		quantity, _ := strconv.ParseFloat(v.Quantity, 64)
		price, _ := strconv.ParseFloat(v.Price, 64)
		quoteQuantity := quantity * price
		table.Append([]string{v.Symbol, side, v.Price, v.CommissionAsset, v.Commission, v.Quantity, fmt.Sprintf("%f", quoteQuantity), time.UnixMilli(v.Time).Format("2006-01-02 15:04:05 -0700")})
	}
	table.Render()

}
