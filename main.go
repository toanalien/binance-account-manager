package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

func avgPrice(asset binance.IsolatedMarginAsset, isolatedMarginTrades []*binance.TradeV3) (avgPrice float64) {
	baseAssetNet, _ := strconv.ParseFloat(asset.BaseAsset.NetAsset, 64)
	calBaseAssetNet := baseAssetNet
	cost := 0.0
	for _, trade := range isolatedMarginTrades {
		if calBaseAssetNet < 0.001 {
			break
		} else {
			quantity, _ := strconv.ParseFloat(trade.Quantity, 64)
			commission, _ := strconv.ParseFloat(trade.Commission, 64)
			price, _ := strconv.ParseFloat(trade.Price, 64)
			if trade.IsBuyer {
				totalQuantity := quantity - commission
				if calBaseAssetNet > totalQuantity {
					calBaseAssetNet -= totalQuantity
				} else {
					calBaseAssetNet = 0
				}
				cost += totalQuantity * price
			}
		}
	}
	avgPrice = cost / baseAssetNet
	return
}

func main() {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()
	var (
		apiKey         = viper.Get("API_KEY").(string)
		secretKey      = viper.Get("SECRET_KEY").(string)
		discordWebHook = viper.Get("DISCORD_WEBHOOK").(string)
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
	table.SetHeader([]string{"Symbol", "QuoteAsset", "BaseAsset", "LiquidatePrice", "IndexPrice", "Balance"})
	for _, v := range userAssets {
		liquidatePrice, _ := strconv.ParseFloat(v.LiquidatePrice, 64)
		quoteAsset, _ := strconv.ParseFloat(v.QuoteAsset.NetAsset, 64)
		baseAsset, _ := strconv.ParseFloat(v.BaseAsset.NetAsset, 64)
		indexPrice, _ := strconv.ParseFloat(v.IndexPrice, 64)
		balance := baseAsset*indexPrice + quoteAsset
		if liquidatePrice != 0 || quoteAsset != 0 || baseAsset != 0 {
			table.Append([]string{v.Symbol, v.QuoteAsset.NetAsset, v.BaseAsset.NetAsset, v.LiquidatePrice, v.IndexPrice, fmt.Sprintf("%f", balance)})
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
	table.SetHeader([]string{"Symbol", "ExecutedQuantity", "OrigQuantity", "Side", "Price", "Cost", "Status", "Type", "TimeInForce"})
	for _, v := range isolatedMarginOpenOrders {
		origQuantity, _ := strconv.ParseFloat(v.OrigQuantity, 64)
		price, _ := strconv.ParseFloat(v.Price, 64)
		cost := origQuantity * price
		table.Append([]string{v.Symbol, v.ExecutedQuantity, v.OrigQuantity, string(v.Side), v.Price, fmt.Sprintf("%.2f", cost), string(v.Status), string(v.Type), time.UnixMilli(v.Time).Format("2006-01-02 15:04:05 -0700")})
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

	for _, v := range userAssets {
		//table = tablewriter.NewWriter(os.Stdout)
		buf := new(bytes.Buffer)
		table = tablewriter.NewWriter(buf)

		if v.Symbol == "BTCUSDT" {
			indexPrice, _ := strconv.ParseFloat(v.IndexPrice, 64)
			baseAssetNet, _ := strconv.ParseFloat(v.BaseAsset.NetAsset, 64)
			liquidatePrice, _ := strconv.ParseFloat(v.LiquidatePrice, 64)
			avgPrice := avgPrice(v, isolatedMarginTrades)
			profit := baseAssetNet * (indexPrice - avgPrice)
			liquidatedRate, _ := strconv.ParseFloat(v.LiquidateRate, 64)
			table.SetHeader([]string{"BTCUSDT"})
			table.Append([]string{"Avg", fmt.Sprintf("%.2f", avgPrice)})
			table.Append([]string{"Index", fmt.Sprintf("%.2f", indexPrice)})
			table.Append([]string{"Profit $", fmt.Sprintf("%.2f", profit)})
			table.Append([]string{"Liquidated", fmt.Sprintf("%.2f", liquidatePrice)})
			table.Append([]string{"% To Liq", fmt.Sprintf("%.2f", liquidatedRate)})
			table.Render()

			content := map[string]string{"content": "```" + buf.String() + "```"}
			jsonValue, _ := json.Marshal(content)
			req, err := http.NewRequest("POST", discordWebHook, bytes.NewBuffer(jsonValue))
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			do, err := client.Do(req)
			_ = do
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
