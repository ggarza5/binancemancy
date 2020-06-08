package main

import (
    "fmt"
    "net/http"
    "github.com/ggarza5/go-binance-futures"
    "github.com/pborman/getopt/v2"
    "context"
    "log"
    "html/template"
    "os"
    "path/filepath"
    "math"
    "strconv"
    _ "strings"
)

var (
    apiKey = "jIyd39L4YfD5CRvygwh5LY1IVilQ38NXY5RshUxKGwR1Sjj6ZGzynkxfK1p2jX0c"
    secretKey = "3IbVAdTpwMN417BNbiwxc63NMpm0EZiBRbC7YFol4gbMytV4FxtfBfJ5dGkgq5Z2"
    openDirection = "SELL"
    tradeDirection = ""
    globalOffset = 0.01
    positionMultiplier = 0.1

    globalOffsetFlag = "0.01"
    positionMultiplierFlag = "0.1"
    mode = "market"
    //TODO: Add betaCorrectPositions
    //Check get last batch of orders, and  for those that have not triggered, perform a market order to fill the position at the current price of the ordered asset
    //Another variant - transfer margin from the high beta assets whose orders have triggered to the low beta ones that didnt trigger on the drawdown
    //TODO: check price differential from high to low and start of order period to end of period/experiment with different time frames and methodologies for beta calculation

    //TODO: user interface

    //TODO: ADd GET /fapi/v1/positionRisk to account service and
    // DELETE /fapi/v1/allOpenOrders to order service

/*  Limit orders being placed -- TRX, XRP, BCH, XLM,ADA,XMR,ATOM
     others are NOT BEING PLACED!!!!! instantly
    account now is 1450
    //targetting 6 positions using 50% of margin
    //15 assets
    //right now this is $35 for a tenth of a position
    // positionSizes = []float64{0.5,750,45,3,11500,17.5,75,3000,3000,3,2,4,57,25}
    //current position open = 0
    */
    client = binance.NewClient(apiKey, secretKey)
    pairs = []string{"BCHUSDT", "XRPUSDT", "EOSUSDT", "LTCUSDT", "TRXUSDT", "ETCUSDT", "LINKUSDT", "XLMUSDT", "ADAUSDT", "XMRUSDT", "DASHUSDT", "ZECUSDT", "XTZUSDT", "BNBUSDT", "ATOMUSDT", "ONTUSDT", "IOTAUSDT", "BATUSDT", "VETUSDT", "NEOUSDT", "THETAUSDT"}
    //I want to be able to have 3 positions take half of my margin
    //for testing 
    positionSizes = map[string]float64{
        "BCHUSDT":0.5,
        "XRPUSDT":1100,
        "EOSUSDT":80,
        "LTCUSDT":3,
        "TRXUSDT":11500,
        "ETCUSDT":20,
        "LINKUSDT":60,
        "XLMUSDT":3000,
        "ADAUSDT":3000,
        "XMRUSDT":1.75,
        "DASHUSDT":2,
        "ZECUSDT":4,
        "XTZUSDT":57,
        "ATOMUSDT":50,
        "BNBUSDT":6,
        "ONTUSDT":200,
        "IOTAUSDT":1400,
        "BATUSDT":1000,
        "VETUSDT":25000,
        "NEOUSDT":12,
        "QTUMUSDT":100,
        "IOSTUSDT":20000,
        "THETAUSDT":800,
    }

    openPositions = map[string]float64{
        "BCHUSDT":0,
        "XRPUSDT":0,
        "EOSUSDT":0,
        "LTCUSDT":0,
        "TRXUSDT":0,
        "ETCUSDT":0,
        "LINKUSDT":0,
        "XLMUSDT":0,
        "ADAUSDT":0,
        "XMRUSDT":0,
        "DASHUSDT":0,
        "ZECUSDT":0,
        "XTZUSDT":0,
        "ATOMUSDT":0,
        "BNBUSDT":0,
        "ONTUSDT":0,
        "IOTAUSDT":0,
        "BATUSDT":0,
        "VETUSDT":0,
        "NEOUSDT":0,
        "QTUMUSDT":0,
        "IOSTUSDT":0,
        "THETAUSDT":0,        
    }

    pricePrecisions = map[string]int{
        "BCHUSDT":2,
        "XRPUSDT":4,
        "EOSUSDT":3,
        "LTCUSDT":2,
        "TRXUSDT":5,
        "ETCUSDT":3,
        "LINKUSDT":3,
        "XLMUSDT":5,
        "ADAUSDT":5,
        "XMRUSDT":2,
        "DASHUSDT":2,
        "ZECUSDT":2,
        "XTZUSDT":3,
        "ATOMUSDT":3,
        "BNBUSDT":3,
        "ONTUSDT":4,
        "IOTAUSDT":4,
        "BATUSDT":4,
        "VETUSDT":6,
        "NEOUSDT":3,
        "QTUMUSDT":3,
        "IOSTUSDT":6,
        "THETAUSDT":4,     
    } 

    positionPrecisions = map[string]int{
        "BCHUSDT":3,
        "XRPUSDT":1,
        "EOSUSDT":1,
        "LTCUSDT":3,
        "TRXUSDT":0,
        "ETCUSDT":2,
        "LINKUSDT":2,
        "XLMUSDT":0,
        "ADAUSDT":0,
        "XMRUSDT":3,
        "DASHUSDT":3,
        "ZECUSDT":3,
        "XTZUSDT":1,
        "ATOMUSDT":2,
        "BNBUSDT":2,
        "ONTUSDT":1,
        "IOTAUSDT":1,
        "BATUSDT":1,
        "VETUSDT":0,
        "NEOUSDT":2,
        "QTUMUSDT":1,
        "IOSTUSDT":0,
        "THETAUSDT":1,   
    }              

    // openPositions = []float64{0,0,0,0,0,0,0,0,0,0,0,0,0,0}
    // pricePrecisions = []int{2,4,3,2,5,3,3,5,5,2,2,2,3,3}
    // positionPrecisions = []int{3,1,1,3,0,2,2,0,0,3,3,3,1,2}

)

//begin getopt initialization 

/*
 * function init
 * params: 
 ************************
 * Initiates the global flag variables
 */
func init() {
    getopt.FlagLong(&tradeDirection, "dir", 'd', "direction").SetOptional()
    getopt.FlagLong(&positionMultiplierFlag, "mult", 'm', "multiplier").SetOptional()
    getopt.FlagLong(&globalOffsetFlag, "off", 'o', "off").SetOptional()
    getopt.FlagLong(&mode, "mode", 'M', "mode").SetOptional()
}

func getAccount(client *binance.Client) *binance.Account {
    account, _  := client.NewGetAccountService().Do(context.Background())
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
    println(tradeDirection)
    println(globalOffsetFlag)
    println(positionMultiplierFlag)
    println(mode)
}

//TODO:genericize cleaning of arguments using "reflect" package and interface pointers
func cleanFlagArguments() {
    if len(tradeDirection) == 0 {
        tradeDirection = "cancel"      
    }  else if tradeDirection[0] == ' ' || tradeDirection[0] == '=' {
        tradeDirection = tradeDirection[1:]
    }
    if tradeDirection[0] == 's' || tradeDirection[0] == 'S' {
        tradeDirection = "SELL"
    } else if tradeDirection[0] == 'l' || tradeDirection[0] == 'L' {
        tradeDirection = "BUY"
        println(tradeDirection)        
    } else {
        if tradeDirection != "cancel" {
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
    } else if mode[0] == 'l'|| mode[0] == 'L' {
        mode = "limit"
    }  else if mode[0] == 'c'|| mode[0] == 'C' {
        mode = "cancel"
    } else if mode[0] == 's'|| mode[0] == 'S' {
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

func listenOnSocket(client *binance.Client, pair string, timeframe string) {
    wsKlineHandler := func(event *binance.WsKlineEvent) {
        if event.Kline.IsFinal {
            fmt.Println(event)
            ltcFifteen = append(ltcFifteen, event)
            // marketOrders(client, )
        }
        // fmt.Println(ltcFifteen)
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
    println("jhhh")
    getopt.Parse()
    cleanFlagArguments()
    // printFlagArguments()
    client := binance.NewClient(apiKey, secretKey)
    client.Debug = true
    if mode == "market" {
        marketOrders(client, stringToSide(tradeDirection), pairs...)
    } else if mode == "limit" {
        limitOrders(client, stringToSide(tradeDirection), globalOffset, pairs...)
    } else if mode == "account" {
        account, _  := client.NewGetAccountService().Do(context.Background())
        println(account)        
    } else if mode == "cancel" {
        cancelOrders(client, getOrders(client))
    } else if mode == "close" {
        closeOpenPositions(client)
    } else if mode == "server" {
        // setupServer()
        listenOnSocket(client, "BTCUSDT", "1m")

    } else if mode == "stopMarket" {
        //longs will have stops set below current prices,
        //shorts will have stops set above current prices
        stopMarketOrders(client, stringToSide(tradeDirection), globalOffset, binance.OrderReduceOnlyFalse, pairs...)
    } else if mode == "stopMarketReduceOnly" {
        //longs will have stops set below current prices,
        //shorts will have stops set above current prices
        stopMarketOrders(client, stringToSide(tradeDirection), globalOffset, binance.OrderReduceOnlyTrue, pairs...)
    } else {
        log.Fatal("Utility not called with a suitable argument to the mode flag. Exiting without execution.")
    }
}

func setupServer() {
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.HandleFunc("/", serveTemplate)
    http.HandleFunc("/market_buy.json", marketBuyHandler)
    http.HandleFunc("/market_sell.json", marketSellHandler)
    http.HandleFunc("/limit_buy.json", limitBuyHandler)
    http.HandleFunc("/limit_sell.json", limitSellHandler)
    http.HandleFunc("/close.json", closeHandler)
    http.HandleFunc("/cancel.json", cancelHandler)
    log.Println("Listening...")
    http.ListenAndServe(":8080", nil)
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
    lp := filepath.Join("templates", "layout.html")
    fp := filepath.Join("templates", filepath.Clean(r.URL.Path))
    // println(r.URL.Path)
    // Return a 404 if the template doesn't exist
    info, err := os.Stat(fp)
    if err != nil {
        if os.IsNotExist(err) {
          http.NotFound(w, r)
          return
        }
    }

    // Return a 404 if the request is for a directory
    if info.IsDir() {
        http.NotFound(w, r)
        return
    }

    tmpl, err := template.ParseFiles(lp, fp)
    if err != nil {
        // Log the detailed error
        log.Println(err.Error())
        // Return a generic "Internal Server Error" message
        http.Error(w, http.StatusText(500), 500)
        return
    }

    if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
        log.Println(err.Error())
        http.Error(w, http.StatusText(500), 500)
    }
}

func marketBuyHandler(w http.ResponseWriter, r *http.Request) {
    tradeDirection = "BUY"    
    // marketOrders()
    marketOrders(client, stringToSide(tradeDirection))    
    // openDirection = "BUY"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func marketSellHandler(w http.ResponseWriter, r *http.Request) {
    tradeDirection = "SELL"
    // marketOrders()
    marketOrders(client, stringToSide(tradeDirection))        
    openDirection = "SELL"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func limitBuyHandler(w http.ResponseWriter, r *http.Request) {
    tradeDirection = "BUY"    
    // limitOrders()
    limitOrders(client, stringToSide(tradeDirection), globalOffset)    
    openDirection = "BUY"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func limitSellHandler(w http.ResponseWriter, r *http.Request) {
    tradeDirection = "SELL"    
    // marketOrders()
    limitOrders(client, stringToSide(tradeDirection), globalOffset)    
    openDirection = "SELL"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func cancelHandler(w http.ResponseWriter, r *http.Request) {
    cancelOrders(client, getOrders(client))    
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func closeHandler(w http.ResponseWriter, r *http.Request) {
    closeOpenPositions(client)
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
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
    for _, asset := range pairs {
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
    fmt.Println(prices)
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
 * function HelloServer
 * params: client
 ************************
 */
func HelloServer(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!<br><button><a href='/'>Buy</a></button><br><button><a href='/'>Sell</a></button><br><button><a href='/'>Close</a></button><br>", r.URL.Path[1:])
}

/*
 * function closeOpenPositions
 * params: client
 ************************
 *///Assumes 
func closeOpenPositions(client *binance.Client) {
    for _, asset := range pairs {
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
    if openPositions[asset] > 0 {
        err = marketOrder(client, "SELL", asset, fmt.Sprintf("%f", openPositions[asset]))
    } else if openPositions[asset] < 0 {
        err = marketOrder(client, "BUY", asset, fmt.Sprintf("%f", -1 * openPositions[asset]))        
    }
    if err == nil {
        openPositions[asset] = 0
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
    if (positionPrecisions[asset] == 1) { 
        return fmt.Sprintf("%.1f", size)
    } else if (positionPrecisions[asset] == 2) { 
        return fmt.Sprintf("%.2f", size)
    } else if (positionPrecisions[asset] == 3) {
        return fmt.Sprintf("%.3f", size)
    } else if (positionPrecisions[asset] == 4) { 
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
    if (direction == binance.SideTypeBuy) { 
        priceOffset = price * offset 
    } else { 
        priceOffset = price * offset * -1 
    }    
    if (pricePrecisions[asset] == 2) {         
        return fmt.Sprintf("%.2f", price - priceOffset)
    } else if (pricePrecisions[asset] == 3) {        
        return fmt.Sprintf("%.3f", price - priceOffset)
    } else if (pricePrecisions[asset] == 4) {         
        return fmt.Sprintf("%.4f", price - priceOffset)
    } else {         
        return fmt.Sprintf("%.5f", price - priceOffset) 
    }            
}

/*
 * function limitOrders
 * params: client
 ************************
 */
func limitOrders(client *binance.Client, direction binance.SideType, offset float64, pairs ...string) {
    prices := getPrices(client)
    for globalPairIndex, _ := range pairs {
        reversedIndex := len(pairs) - globalPairIndex - 1
        size := calculateOrderSizeFromPrecision(positionMultiplier * positionSizes[pairs[reversedIndex]], pairs[reversedIndex])
        println(reversedIndex)
        println(pairs[reversedIndex])
        fmt.Println(prices)           
        println(prices[reversedIndex])     
        limitOrder(client, direction, pairs[reversedIndex], offset, size, calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, pairs[reversedIndex]))
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
    println("THE REDUCE ONLY STATUS")  
    println(isReduceOnly)
    for globalPairIndex, _ := range pairs {
        reversedIndex := len(pairs) - globalPairIndex - 1        
        size := calculateOrderSizeFromPrecision(positionMultiplier * positionSizes[pairs[reversedIndex]], pairs[reversedIndex])        
        stopMarketOrder(client, getOppositeDirection(direction), pairs[reversedIndex], offset, size, calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, pairs[reversedIndex]), isReduceOnly)
    }
}

/*
 * function stopMarketOrder
 * params: client
 ************************
 *///error-returning
func stopMarketOrder(client *binance.Client, direction binance.SideType, asset string, offset float64, size string, price string, isReduceOnly binance.OrderReduceOnly) error {
    order, err := client.NewCreateOrderService().Symbol(asset).
        Side(direction).Type(binance.OrderTypeStopLoss).
        Quantity(size).ReduceOnly(isReduceOnly).
        StopPrice(price).Do(context.Background())
    if err != nil {
        fmt.Println(err)
        return err
    }
    openPositions[asset] -= positionMultiplier * positionSizes[asset]        
    fmt.Println(order)
    return nil
}
/*
 * function marketOrders
 * params: client
 ************************
 */
func marketOrders(client *binance.Client, direction binance.SideType, pairs ...string) {
    for _, asset := range pairs {
        println(calculateOrderSizeFromPrecision(positionMultiplier * positionSizes[asset], asset))
        marketOrder(client, direction, asset, calculateOrderSizeFromPrecision(positionMultiplier * positionSizes[asset], asset))
    }
}

/*
 * function marketOrder
 * params: client
 ************************
 *///error-returning
func marketOrder(client *binance.Client, direction binance.SideType, asset string, size string) error {
    order, err := client.NewCreateOrderService().Symbol(asset).
        Side(direction).Type(binance.OrderTypeMarket).
        Quantity(size).Do(context.Background())
    if err != nil {
        fmt.Println(err)
        return err
    }
    openPositions[asset] -= positionMultiplier * positionSizes[asset]        
    fmt.Println(order)
    return nil
}
