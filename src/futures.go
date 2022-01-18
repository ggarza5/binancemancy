package main

import (
    "context"
    "fmt"
    discordGo "github.com/bwmarrin/discordgo"

    "github.com/ggarza5/go-binance-futures"
    "github.com/ggarza5/go-binance-futures/common"
    "github.com/ggarza5/technical-indicators"
    "github.com/pborman/getopt/v2"
    "log"
    "os"
    gCommon "github.com/ggarza5/alpaca-first/common"
    pretty "github.com/inancgumus/prettyslice"
    "math"
    _ "math/rand"
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


/*
 * function getAccount
 * params:
 ************************
 * Gets our account object from binance.
 */
func getAccount(client *binance.Client) *binance.Account {
    account, _ := client.NewGetAccountService().Do(context.Background())
    return account
}


/*
 * TODO
 * function init
 * params:
 ************************
 * Initiates the global flag variables
 */
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


/*
 * function parseCommandFlags
 * params: parse
 ************************
 * Parses the flags passed to the command on the console and sets the execution mode 
 * of the libray.
 */
//TODO:genericize cleaning of arguments using "reflect" package and interface pointers

func parseCommandFlags() {
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

/*
 * *********************************************************************
 * ***********CONSTANTS for Indicator Calculation & Handling************
 * *********************************************************************
*/

var globalKlines = make([]*binance.WsKlineEvent, 0)
var globalKlineIndex = 0
var globalKlineRsi = 0.0

var dataCloses = make([]string, 0)
var dataClosesInts = make([]int, 0)
var dataClosesFloats = make([]float64, 0)

var dataOpens = make([]string, 0)
var dataOpensInts = make([]int, 0)
var dataOpensFloats = make([]float64, 0)

var dataHighs = make([]string, 0)
var dataHighsInts = make([]int, 0)
var dataHighsFloats = make([]float64, 0)

var dataLows = make([]string, 0)
var dataLowsInts = make([]int, 0)
var dataLowsFloats = make([]float64, 0)

var dummyCloses []float64 = []float64{1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 4.0, 1.0, 3.0, 1.0, 2.0, 3.0, 1.0, 4.0, 3.0, 10.0, 1.0, 1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 5.0, 1.0, 3.0, 1.0, 1.0, 2.0, 1.0, 1.0, 1.0, 10.0, 1.0, 1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 4.0, 2.0, 1.0, 2.0, 1.0, 3.0, 1.0, 3.0, 3.0, 1.0, 4.0, 1.0, 13.0, 1.0, 4.0, 3.0, 10.0, 1.0, 1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 5.0, 1.0, 3.0, 1.0, 1.0, 2.0}
var dummyHighs []float64 = []float64{2.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 2.0, 1.0, 3.0, 1.0, 1.0, 3.0, 1.0, 4.0, 1.0, 10.0, 1.0}
var dummyLows []float64 = []float64{1.0, 2.0, 2.5, 5.0, 1.0, 2.0, 3.0, 10.0, 10.0, 10.0, 2.0, 1.0, 2.0, 1.0, 3.0, 1.0, 1.0, 3.0, 1.0, 4.0, 1.0, 10.0, 1.0}

var kijunInterval int = 60
var tenkanInterval int = 20
var sks1Interval int = 60
var sks2Interval int = 120

var discordAuthToken = "ODc4ODU0ODY2MjI5OTQ0MzQw.YSHPYA.EJuBjlGTe1VMXrHTFc55bWv2FlY"
var staffChannelId = "719691307983044671"
var testSharkPrivId = "717509948489203823"


/*
 * function printFlagArguments
 * params:
 ************************
 * Logs the length of a slice, then pops the slice if beyond our specified time series length
 *  - used for indictator time series tracking
 */
func popFloatSliceAndLog(s []float64, slen int) []float64 {
    println("the length of the slice is ")
    println(len(s))
    if len(s) > slen {
        println("should be popping")
        return s[1:]
    } else {
        println("we not going to pop")
        return s
    }
}

/*
 * function popFloatSlice
 * params: slice ([]float64, specified legnth (int)
 ************************
 * Pops a slice if it's beyond our specified time series length
 */
func popFloatSlice(s []float64, slen int) []float64 {   
    if len(s) > slen {
        return s[1:]
    } else {
        return s
    }
}


/*
 * function printCloud
 * params:
 ************************
 * Pops a slice if it's beyond our specified time series length
 */
func printCloud(a ...[]float64) {
    for i, x := range a {
        pretty.Show("cloud line "+strconv.Itoa(i), x)
    }
}

//need to do nothing when data arrays are 0
//need to not issue any signals until senkou span b is finished forming
//for now need to wait until

//first edition
//only calculate indicator on candle closes

/*
 * function listenOnSocket
 * params: client *binance.Client, pair string, timeframe string, sess *discordGo.Session
 ************************
 * Deines a handler to listen to signal events coming from the price eolution on binance. Handles the changes in price in a control flow
 */
func listenOnSocket(client *binance.Client, pair string, timeframe string, sess *discordGo.Session) {
    wsKlineHandler := func(event *binance.WsKlineEvent) {
        if !event.Kline.IsFinal {
            return
        }
        fmt.Println(event)
        println("event is final")
        globalKlines = append(globalKlines, event)
        // dataCloses = popFloatSlice(append(dataCloses, event.Kline.Close))
        dataClosesFloats = popFloatSlice(append(dataClosesFloats, gCommon.ParseFloatHandleErr(event.Kline.Close)), sks2Interval)
        dataHighsFloats = popFloatSlice(append(dataHighsFloats, gCommon.ParseFloatHandleErr(event.Kline.High)), sks2Interval)
        // dataHighs = popFloatSlice(append(dataHighs, event.Kline.High))
        // dataLows = popFloatSlice(append(dataLows, event.Kline.Low))
        dataLowsFloats = popFloatSlice(append(dataLowsFloats, gCommon.ParseFloatHandleErr(event.Kline.Low)), sks2Interval)
        // i, _ := strconv.Atoi(event.Kline.Close)
        fmt.Println(dataClosesFloats)

        //calculate cloud and get signal
        conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan := indicators.CalculateIchimokuCloud(dataClosesFloats, dataLowsFloats, dataHighsFloats, []int{9, 26, 52, 26})
        // println(conversionLine)
        // fmt.Println(fmt.Printf("%.2f", conversionLine))
        _, _, _, _, _ = conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan
        printCloud(conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan)
        signal := indicators.DetermineCloudSignal(dataClosesFloats, conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan)
        println(signal)
        if signal == -1 {
            sess.ChannelMessageSend(testSharkPrivId, pair+" just had a negative TK cross on the "+timeframe+" timeframe.")
        } else if signal == 1 {
            sess.ChannelMessageSend(testSharkPrivId, pair+" just had a positive TK cross on the "+timeframe+" timeframe.")
        }
        // if signal == 0 {
        //     sess.ChannelMessageSend(testSharkPrivId, pair+" just had NO CROSS ! "+timeframe+" timeframe.")
        // }

        // }
       
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

//TODO
/*
 * func BollingerBands(data)
 * Calculates the Bollinger Bands of a price series, which is simply the 2 standard deviation range of the
 * time series distribution
*/
func BollingerBands() {}
 //start calculating Bollinger Bands
        /*if len(dummyCloses) >= 20 {
            middle, upper, lower := indicators.BollingerBands(dummyCloses, 20, 2.0)
            _, _, _ = middle, upper, lower
            var dummyHighs []float64
            var dummyLows []float64
            for _, f := range dummyCloses {
                dummyHighs = append(dummyHighs, f+rand.Float64())
                dummyLows = append(dummyLows, f-rand.Float64())
            }

            conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan := indicators.CalculateIchimokuCloud(dummyCloses, dummyLows, dummyHighs, []int{20, 60, 120, 30})
            // println(conversionLine)
            fmt.Println(fmt.Printf("%.2f", conversionLine))
            println("sasda")
            _, _, _, _, _ = conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan
            indicators.DetermineCloudSignal()
        }*/

/*
 * function stringSlicetoFloatSlice
 * params: string slice
 * returns: float slice
 ************************
 * Converts a slice of strings to a slice of floats
 */
func stringSlicetoFloatSlice(s []string) []float64 {
    sliceScan = scanner.Text()
    newSlice := make([]float64, len(s), len(s))
    for i := 0; i < len(s); i += 1 {
        f64, err := strconv.ParseFloat(s[i], 64)
        newSlice[i] = float64(f64)
    }
}

/*
 Current flow CLI utility cold with server mode enabled
 Listen on socket called

*/
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
    parseCommandFlags()
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
        dg, err := discordGo.New("Bot " + discordAuthToken)
        if err != nil {
            fmt.Println("Error creating Discord session: ", err)
            return
        }
        // Spawns goRoutines to listen to multiple different pairs at once from Binance.
        go listenOnSocket(client, "BTCUSDT", "1m", dg)
        // go listenOnSocket(client, "ETHBTC", "1m", dg)
        // go listenOnSocket(client, "BNBBTC", "1m", dg)
        // go listenOnSocket(client, "FTMUSDT", "1m", dg)
        // go listenOnSocket(client, "FTMBTC", "1m", dg)
        // go listenOnSocket(client, "ADABTC", "1m", dg)
        // go listenOnSocket(client, "LINKBTC", "1m", dg)
        exit := make(chan string)
        for {
            select {
            case <-exit:
                os.Exit(0)
            }
        }
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

/*
 * function cancelOrders
 * params: client, order slice
 ************************
 * Cancels a defined set of orders that we pass into the function
 */
func cancelOrders(client *binance.Client, orders []*binance.Order) {
    for _, order := range orders {
        _, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}

/*
 * function cancelOrders
 * params: client
 ************************
 * Cancels all of our outstanding orders on Biannce
 */
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

/*
 * function getOrders
 * params: client
 ************************
 * Gets all of our outstanding orders on Binance
 */
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
 * Gets the price state object from Binance -- Prices of all assets
 * Crrently ignoring the price of BTC and ETH
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
    //skip btc and eth prices, because we're trading altcoins
    return prices[2:]
}

/*
 * function stringToSide
 * params: direction
 ************************
 * Converts a string into a Binance side direction
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
 * Closes all open positions on binance
 */ 
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
 * Places stopMarket orders at a passed differetial from the current price (the offset paraeter)
 * on the  passed pairs and in the specified direction
 * Utilizes stopMarketOrder function
 * The size that will be used by thes market orders will be proportional to the position's total value, 
 * with the proportion being passed into /futuers at the command line invocation. It needs
 * to be calculated by handling the precision used by th pai on Binance, for which it calls 
 * calculateOrderSizeFromPrecision
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
 * 
  * //error-returning, needs to be handled
 * Places orders at the market - at the best available price -
 * on the  passed pairs and in the specified direction
 * Utilizes arketOrder function
 * The size that will be used by thes market orders is set at invocation
 * Logging currently enabled, TODO disable or allow configuration of logging
 */ 
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
 * Places orders at the market - at the best available price -
 * on the  passed pairs and in the specified direction
 * Utilizes arketOrder function
 * The size that will be used by thes market orders is set at invocation
 * Logging currently enabled, TODO disable or allow configuration of logging
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
 * Called by MarketOrdes
 * Places a single order at the market on the passed pair, in the specified sizes and direction
 * //error-returning, needs to be handled
 */ 
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
