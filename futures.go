package main

import (
    "context"
    "fmt"
    "github.com/ggarza5/go-binance-futures"
    // "github.com/adshao/go-binance"
    "github.com/ggarza5/go-binance-futures/common"
    "github.com/ggarza5/technical-indicators"
    "github.com/pborman/getopt/v2"
    "log"
    // "html/template"
    // "os"
    // "path/filepath"
    "math"
    "math/rand"
    "strconv"
    _ "strings"
)

type mfloat []float64

var (
    OpenDirection      = "SELL"
    TradeDirection     = ""
    globalOffset       = 0.01
    positionMultiplier = 0.1

    globalOffsetFlag       = "0.01"
    positionMultiplierFlag = "0.1"
    mode                   = "market"
    //TODO: Add betaCorrectPositions
    //Check get last batch of orders, and  for those that have not triggered, perform a market order to fill the position at the current price of the ordered asset
    //Another variant - transfer margin from the high beta assets whose orders have triggered to the low beta ones that didnt trigger on the drawdown
    //TODO: check price differential from high to low and start of order period to end of period/experiment with different time frames and methodologies for beta calculation

    //TODO: user interface

    //TODO: ADd GET /fapi/v1/positionRisk to account service and
    // DELETE /fapi/v1/allOpenOrders to order service
    client = binance.NewClient(common.ApiKey, common.SecretKey)
)

func Index(vs []string, t string) int {
    for i, v := range vs {
        if v == t {
            return i
        }
    }
    return -1
}

func Includes(vs []string, t string) bool {
    return Index(vs, t) >= 0
}

//begin getopt initialization

/*
 * function init
 * params:
 ************************
 * Initiates the global flag variables
 */
func init() {
    getopt.FlagLong(&TradeDirection, "dir", 'd', "direction").SetOptional()
    getopt.FlagLong(&positionMultiplierFlag, "mult", 'm', "multiplier").SetOptional()
    getopt.FlagLong(&globalOffsetFlag, "off", 'o', "off").SetOptional()
    getopt.FlagLong(&mode, "mode", 'M', "mode").SetOptional()
}

func getAccount(client *binance.Client) *binance.Account {
    account, _ := client.NewGetAccountService().Do(context.Background())
    return account
}

func getPositions(*binance.Account) {
    // account.Positions
}

/*
 * function printFlagArguments
 * params:
 ************************
 * Prints the global flag variables. Should be called after they are set by init()
 */
func printFlagArguments() {
    println(TradeDirection)
    println(globalOffsetFlag)
    println(positionMultiplierFlag)
    println(mode)
}

//TODO:genericize cleaning of arguments using "reflect" package and interface pointers
func cleanFlagArguments() {
    if len(TradeDirection) == 0 {
        TradeDirection = "cancel"
    } else if TradeDirection[0] == ' ' || TradeDirection[0] == '=' {
        TradeDirection = TradeDirection[1:]
    }
    if TradeDirection[0] == 's' || TradeDirection[0] == 'S' {
        TradeDirection = "SELL"
    } else if TradeDirection[0] == 'l' || TradeDirection[0] == 'L' {
        TradeDirection = "BUY"
        println(TradeDirection)
    } else {
        if TradeDirection != "cancel" {
            log.Fatal("Direction flag used with an unsuitable argument.")
        }
    }
    if globalOffsetFlag[0] == ' ' || globalOffsetFlag[0] == '=' {
        globalOffsetFlag = globalOffsetFlag[1:]
    }
    globalOffset, _ = strconv.ParseFloat(globalOffsetFlag, 64)
    if positionMultiplierFlag[0] == ' ' || positionMultiplierFlag[0] == '=' {
        positionMultiplierFlag = positionMultiplierFlag[1:]
    }
    positionMultiplier, _ = strconv.ParseFloat(positionMultiplierFlag, 64)
    if mode[0] == ' ' || mode[0] == '=' {
        mode = mode[1:]
    }
    if mode[0] == 'm' || mode[0] == 'M' {
        mode = "market"
    } else if mode[0] == 'l' || mode[0] == 'L' {
        mode = "limit"
    } else if mode[0] == 'c' || mode[0] == 'C' {
        mode = "cancel"
    } else if mode[0] == 's' || mode[0] == 'S' {
        if len(mode) < 2 {
            mode = "server"
        } else if len(mode) < 3 {
            mode = "stopMarket"
        } else if len(mode) < 4 {
            println("we got to reduce only!!!")
            mode = "stopMarketReduceOnly"
        } else {
            mode = "server"
        }
    } else {
        // r, _ := client.NewPremiumIndexService().Do(context.Background())
        // println(r)
        log.Fatal("Mode flag used with an unsuitable argument.")
    }
}

//TODO
// func shortIfPriceClosedBelowLevel(client *binance.Client, pair string, timeframe string, price string)
//calculate trendlines
//use intercept and slope

//func shortIfPriceClosedBelowTrend

var ltcFifteen = make([]*binance.WsKlineEvent, 0)
var ltcFifteenIndex = 0
var ltcFifteenRSI = 0.0

var dataCloses = make([]string, 0)
var dataClosesInts = make([]int, 0)
var dataClosesFloats = make([]float64, 32)

var dummyCloses []float64 = []float64{1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 4.0, 1.0, 3.0, 1.0, 2.0, 3.0, 1.0, 4.0, 3.0, 10.0, 1.0, 1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 5.0, 1.0, 3.0, 1.0, 1.0, 2.0, 1.0, 1.0, 1.0, 10.0, 1.0, 1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 4.0, 2.0, 1.0, 2.0, 1.0, 3.0, 1.0, 3.0, 3.0, 1.0, 4.0, 1.0, 13.0, 1.0, 4.0, 3.0, 10.0, 1.0, 1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 5.0, 1.0, 3.0, 1.0, 1.0, 2.0}
var dummyHighs []float64 = []float64{2.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 2.0, 1.0, 3.0, 1.0, 1.0, 3.0, 1.0, 4.0, 1.0, 10.0, 1.0}
var dummyLows []float64 = []float64{1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 2.0, 1.0, 3.0, 1.0, 1.0, 3.0, 1.0, 4.0, 1.0, 10.0, 1.0}

func listenOnSocket(client *binance.Client, pair string, timeframe string) {
    wsKlineHandler := func(event *binance.WsKlineEvent) {
        if event.Kline.IsFinal {
            // fmt.Println(event)
            ltcFifteen = append(ltcFifteen, event)
            dataCloses = append(dataCloses, event.Kline.Close)
            i, _ := strconv.Atoi(event.Kline.Close)
            dataClosesInts = append(dataClosesInts, i)
            dataClosesFloats = append(dataClosesFloats, float64(i))
            fmt.Println(dataCloses)
        }
        //start calculating Bollinger Bands
        if len(dummyCloses) >= 20 {
            middle, upper, lower := indicators.BollingerBands(dummyCloses, 20, 2.0)
            _, _, _ = middle, upper, lower
            var dummyHighs []float64
            var dummyLows []float64
            for _, f := range dummyCloses {
                dummyHighs = append(dummyHighs, f+rand.Float64())
                dummyLows = append(dummyLows, f-rand.Float64())
            }

            conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan := indicators.IchimokuCloud(dummyCloses, dummyLows, dummyHighs, []int{20, 60, 120, 30})
            _, _, _, _, _ = conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan
        }
    }
    errHandler := func(err error) {
        fmt.Println(err)
    }
    doneC, _, err := binance.WsKlineServe(pair, timeframe, wsKlineHandler, errHandler)
    if err != nil {
        fmt.Println(err)
        return
    }
    <-doneC
}

//end getopt initialization
//invoke like ./main -dir=short -mult=1.0
/*
 * function main
 * params:
 ************************
 */
func main() {
    println("Starting futures.go")
    getopt.Parse()
    cleanFlagArguments()
    // printFlagArguments()
    client := binance.NewClient(common.ApiKey, common.SecretKey)
    client.Debug = true
    if mode == "market" {
        MarketOrders(client, stringToSide(TradeDirection), common.Pairs...)
    } else if mode == "limit" {
        LimitOrders(client, stringToSide(TradeDirection), globalOffset, common.Pairs...)
    } else if mode == "account" {
        account, _ := client.NewGetAccountService().Do(context.Background())
        println(account)
    } else if mode == "cancel" {
        cancelOrders(client, getOrders(client))
    } else if mode == "close" {
        closeOpenPositions(client)
    } else if mode == "server" {
        // SetupServer()
        println("Running in server mode.")
        listenOnSocket(client, "BTCUSDT", "1m")

    } else if mode == "stopMarket" {
        //longs will have stops set below current prices,
        //shorts will have stops set above current prices
        stopMarketOrders(client, stringToSide(TradeDirection), globalOffset, binance.OrderReduceOnlyFalse, common.Pairs...)
    } else if mode == "stopMarketReduceOnly" {
        //longs will have stops set below current prices,
        //shorts will have stops set above current prices
        stopMarketOrders(client, stringToSide(TradeDirection), globalOffset, binance.OrderReduceOnlyTrue, common.Pairs...)
    } else {
        log.Fatal("Utility not called with a suitable argument to the mode flag. Exiting without execution.")
    }
}

func cancelOrders(client *binance.Client, orders []*binance.Order) {
    for _, order := range orders {
        _, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}

func cancelAllOrders(client *binance.Client) {
    orders := getOrders(client)
    for _, order := range orders {
        _, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}

func getOrders(client *binance.Client) []*binance.Order {
    openOrdersAcrossAllPairs := []*binance.Order{}
    for _, asset := range common.Pairs {
        openOrders, err := client.NewListOpenOrdersService().Symbol(asset).
            Do(context.Background())
        if err != nil {
            fmt.Println(err)
            return nil
        }
        // for _, o := range openOrders {
        //     fmt.Println(o)
        // }
        for _, order := range openOrders {
            openOrdersAcrossAllPairs = append(openOrdersAcrossAllPairs, order)
        }
    }
    return openOrdersAcrossAllPairs
}

/*
 * function getPrices
 * params: client
 ********
 */
func getPrices(client *binance.Client) []*binance.SymbolPrice {
    prices, err := client.NewListPricesService().Do(context.Background())
    if err != nil {
        fmt.Println(err)
        fmt.Println("Error getting prices")
        return nil
    }
    // fmt.Println(prices)
    for _, i := range prices {
        price, _ := strconv.ParseFloat(i.Price, 64)
        if i.Symbol == "BTCUSDT" || i.Symbol == "ETHUSDT" {
            continue
        }
        truncPrice, _ := strconv.ParseFloat(calculateOrderSizeFromPrecision(10/price, i.Symbol), 64)
        fmt.Println(i.Symbol, truncPrice)
    }
    //skip btc and eth prices
    return prices[2:]
}

/*
 * function stringToSide
 * params: direction
 ************************
 */
func stringToSide(direction string) binance.SideType {
    if direction == "BUY" {
        return binance.SideTypeBuy
    } else {
        return binance.SideTypeSell
    }
}

/*
 * function closeOpenPositions
 * params: client
 ************************
 */ //Assumes
func closeOpenPositions(client *binance.Client) {
    for _, asset := range common.Pairs {
        closeOpenPosition(client, asset)
    }
}

/*
 * function closeOpenPosition
 * params: client, pairIndex
 ************************
 */
func closeOpenPosition(client *binance.Client, asset string) {
    var err error
    if common.OpenPositions[asset] > 0 {
        err = marketOrder(client, "SELL", asset, fmt.Sprintf("%f", common.OpenPositions[asset]))
    } else if common.OpenPositions[asset] < 0 {
        err = marketOrder(client, "BUY", asset, fmt.Sprintf("%f", -1*common.OpenPositions[asset]))
    }
    if err == nil {
        common.OpenPositions[asset] = 0
    }
}

func round(x, unit float64) float64 {
    return math.Round(x/unit) * unit
}

/*
 * function calculateOrderSizeFromPrecision
 * params: priceString string, offset float64, direction binance.SideType
 ************************
 */
func calculateOrderSizeFromPrecision(size float64, asset string) string {
    if common.PositionPrecisions[asset] == 0 {
        return fmt.Sprintf("%.0f", size)
    } else if common.PositionPrecisions[asset] == 1 {
        return fmt.Sprintf("%.1f", size)
    } else if common.PositionPrecisions[asset] == 2 {
        return fmt.Sprintf("%.2f", size)
    } else if common.PositionPrecisions[asset] == 3 {
        return fmt.Sprintf("%.3f", size)
    } else if common.PositionPrecisions[asset] == 4 {
        return fmt.Sprintf("%.4f", size)
    } else {
        return fmt.Sprintf("%.5f", size)
    }
}

/*
 * function calculateOrderPriceFromOffset
 * params: priceString string, offset float64, direction binance.SideType
 ************************
 */
func calculateOrderPriceFromOffset(priceString string, offset float64, direction binance.SideType, asset string) string {
    price, err := strconv.ParseFloat(priceString, 64)
    if err != nil {
        fmt.Println(err)
        return ""
    }
    var priceOffset float64
    if direction == binance.SideTypeBuy {
        priceOffset = price * offset
    } else {
        priceOffset = price * offset * -1
    }
    if common.PricePrecisions[asset] == 2 {
        return fmt.Sprintf("%.2f", price-priceOffset)
    } else if common.PricePrecisions[asset] == 3 {
        return fmt.Sprintf("%.3f", price-priceOffset)
    } else if common.PricePrecisions[asset] == 4 {
        return fmt.Sprintf("%.4f", price-priceOffset)
    } else {
        return fmt.Sprintf("%.5f", price-priceOffset)
    }
}

/*
 * function limitOrders
 * params: client
 ************************
 */
func LimitOrders(client *binance.Client, direction binance.SideType, offset float64, pairs ...string) {
    prices := getPrices(client)
    for globalPairIndex := range common.Pairs {
        reversedIndex := len(pairs) - globalPairIndex - 1
        size := calculateOrderSizeFromPrecision(positionMultiplier*common.PositionSizes[pairs[reversedIndex]], common.Pairs[reversedIndex])
        println(reversedIndex)
        println(pairs[reversedIndex])
        fmt.Println(prices)
        println(prices[reversedIndex])
        limitOrder(client, direction, common.Pairs[reversedIndex], offset, size, calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, common.Pairs[reversedIndex]))
    }
}

/*
 * function limitOrder
 * params: client
 ************************
 */
func limitOrder(client *binance.Client, direction binance.SideType, asset string, offset float64, size string, price string) {
    order, err := client.NewCreateOrderService().Symbol(asset).
        Side(direction).Type(binance.OrderTypeLimit).
        TimeInForce(binance.TimeInForceTypeGTC).Quantity(size).
        Price(price).Do(context.Background())
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(order)
}

/*
 * function getOppositeDirection
 * params: client
 ************************
 * Simply returns the opposite side type as the argument. Used for stoplosses
 */
func getOppositeDirection(direction binance.SideType) binance.SideType {
    if direction == binance.SideTypeBuy {
        return binance.SideTypeSell
    } else {
        return binance.SideTypeBuy
    }
}

/*
 * function stopMarketOrders
 * params: client
 ************************
 */
func stopMarketOrders(client *binance.Client, direction binance.SideType, offset float64, isReduceOnly binance.OrderReduceOnly, pairs ...string) {
    prices := getPrices(client)
    for globalPairIndex := range common.Pairs {
        reversedIndex := len(pairs) - globalPairIndex - 1
        size := calculateOrderSizeFromPrecision(positionMultiplier*common.PositionSizes[pairs[reversedIndex]], common.Pairs[reversedIndex])
        stopMarketOrder(client, getOppositeDirection(direction), common.Pairs[reversedIndex], offset, size, calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, common.Pairs[reversedIndex]), isReduceOnly)
    }
}

/*
 * function stopMarketOrder
 * params: client
 ************************
 */ //error-returning
func stopMarketOrder(client *binance.Client, direction binance.SideType, asset string, offset float64, size string, price string, isReduceOnly binance.OrderReduceOnly) error {
    order, err := client.NewCreateOrderService().Symbol(asset).
        Side(direction).Type(binance.OrderTypeStopLoss).
        Quantity(size).ReduceOnly(isReduceOnly).
        StopPrice(price).Do(context.Background())
    if err != nil {
        fmt.Println(err)
        return err
    }
    common.OpenPositions[asset] -= positionMultiplier * common.PositionSizes[asset]
    fmt.Println(order)
    return nil
}

/*
 * function MarketOrders
 * params: client
 ************************
 */
func MarketOrders(client *binance.Client, direction binance.SideType, pairs ...string) {
    for _, asset := range common.Pairs {
        println(calculateOrderSizeFromPrecision(positionMultiplier*common.PositionSizes[asset], asset))
        marketOrder(client, direction, asset, calculateOrderSizeFromPrecision(positionMultiplier*common.PositionSizes[asset], asset))
    }
}

/*
 * function marketOrder
 * params: client
 ************************
 */ //error-returning
func marketOrder(client *binance.Client, direction binance.SideType, asset string, size string) error {
    order, err := client.NewCreateOrderService().Symbol(asset).
        Side(direction).Type(binance.OrderTypeMarket).
        Quantity(size).Do(context.Background())
    if err != nil {
        fmt.Println(err)
        return err
    }
    common.OpenPositions[asset] -= positionMultiplier * common.PositionSizes[asset]
    fmt.Println(order)
    return nil
}
