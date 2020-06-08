package main

import (
    "context"
    "fmt"
    "github.com/ggarza5/go-binance-margin"
    "github.com/pborman/getopt/v2"
    "html/template"
    "log"
    "math"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

//TODO: Option to flip all BTC to Tether and back
//TODO: pre-empt a binance API error of "order size too small", especially on 'close position(with stop)'. Calculate proportion
//of open position before sending any API commands to prevent unstable state of personal order book

var (
    apiKey                 = "jIyd39L4YfD5CRvygwh5LY1IVilQ38NXY5RshUxKGwR1Sjj6ZGzynkxfK1p2jX0c"
    secretKey              = "3IbVAdTpwMN417BNbiwxc63NMpm0EZiBRbC7YFol4gbMytV4FxtfBfJ5dGkgq5Z2"
    openDirection          = "SELL"
    tradeDirection         = ""
    globalOffset           = 0.01
    positionMultiplier     = 1.0
    excludedPairs          = []string{}
    globalOffsetFlag       = "0.01"
    positionMultiplierFlag = "1.0"
    mode                   = "market"

    client = binance.NewClient(apiKey, secretKey)
    //these are the pairs that are not on bitmex, so i would onyl sohrt them on binance
    pairs          = []string{"MATIC", "BNB", "LINK", "ATOM", "DASH", "ZEC", "ONT", "NEO", "XTZ", "ETC", "IOST", "IOTA", "XLM", "RVN", "BAT", "QTUM", "XMR", "VET", "ETH", "LTC"}
    pairsFlag      = ""
    proportion     = 1.0
    proportionFlag = ""
    excludedFlag   = ""
    //current
    accountSize = 0.1
    //TODO: change to calculate these minimumPositionSizes on the fly by pulling prices
    minimumPositionSizes = map[string]float64{
        "MATIC": 100,
        "BNB":   0.05,
        "LINK":  1,
        "LTC":   0.1,
        "ATOM":  0.5,
        "DASH":  0.02,
        "ZEC":   0.02,
        "ONT":   2,
        "NEO":   0.16,
        "XTZ":   0.5,
        "ETC":   0.15,
        "ETH":   0.01,
        "IOST":  240,
        "IOTA":  8,
        "XLM":   30,
        "RVN":   50,
        "BAT":   5,
        "QTUM":  0.6,
        "XMR":   0.02,
        "TRX":   60,
        "VET":   320,
    }
    positionPrecisions = map[string]int{
        "BNB":   2,
        "LINK":  0,
        "ATOM":  2,
        "EOS":   2,
        "DASH":  3,
        "ZEC":   3,
        "ONT":   2,
        "NEO":   2,
        "LTC":   2,
        "MATIC": 0,
        "XTZ":   2,
        "ETC":   2,
        "ETH":   3,
        "IOST":  0,
        "IOTA":  0,
        "XLM":   0,
        "RVN":   0,
        "BAT":   0,
        "QTUM":  2,
        "XMR":   3,
        "TRX":   0,
        "VET":   0,
    }

    pricePrecisions = map[string]int{
        "ZEC":   6,
        "LINK":  8,
        "NEO":   6,
        "ETH":   6,
        "DASH":  6,
        "TRX":   8,
        "LTC":   6,
        "MATIC": 8,
        "XTZ":   7,
        "ETC":   7,
        "IOST":  8,
        "IOTA":  8,
        "XLM":   8,
        "RVN":   8,
        "BNB":   7,
        "BAT":   8,
        "QTUM":  7,
        "VET":   8,
        "ATOM":  7,
        "XMR":   6,
    }
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

func stringsContains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

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
    getopt.FlagLong(&pairsFlag, "pairs", 'p', "pairs").SetOptional()
    getopt.FlagLong(&proportionFlag, "proportion", 'P', "proportion").SetOptional()
    getopt.FlagLong(&excludedFlag, "excluded", 'E', "excluded").SetOptional()
}

func getAccount(client *binance.Client) *binance.Account {
    account, _ := client.NewGetAccountService().Do(context.Background())
    return account
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
    println(pairsFlag)
}

func convertPairsToBinanceUserAssets(pairsArgument []string) []binance.UserAsset {
    assetsToReturn := []binance.UserAsset{}
    for _, p := range pairsArgument {
        userAssetFromPair := binance.UserAsset{
            Asset:    p,
            Borrowed: "",
            Free:     "",
            Interest: "",
            Locked:   "",
            NetAsset: "",
        }
        assetsToReturn = append(assetsToReturn, userAssetFromPair)
    }
    return assetsToReturn
}

//TODO:genericize cleaning of arguments using "reflect" package and interface pointers
func cleanFlagArguments() {  
    printFlagArguments()
    if len(pairsFlag) != 0 {
        if pairsFlag[0] == ' ' || pairsFlag[0] == '=' {
            pairsFlag = pairsFlag[1:]
        }
        pairs = strings.Split(pairsFlag, ",")
    }
    if len(excludedFlag) != 0 {
        if excludedFlag[0] == ' ' || excludedFlag[0] == '=' {
            excludedFlag = excludedFlag[1:]
        }
        excludedPairs = strings.Split(excludedFlag, ",")
    }
    if len(proportionFlag) != 0 {
        if proportionFlag[0] == ' ' || proportionFlag[0] == '=' {
            proportionFlag = proportionFlag[1:]
        }
        proportion, _ = strconv.ParseFloat(proportionFlag, 64)
    }
    if len(tradeDirection) == 0 {
        tradeDirection = "cancel"
    } else if tradeDirection[0] == ' ' || tradeDirection[0] == '=' {
        tradeDirection = tradeDirection[1:]
    }
    if tradeDirection[0] == 's' || tradeDirection[0] == 'S' {
        tradeDirection = "SELL"
    } else if tradeDirection[0] == 'l' || tradeDirection[0] == 'L' {
        tradeDirection = "BUY"
    } else {
        if tradeDirection != "cancel" {
            log.Fatal("Direction flag used with an unsuitable argument.")
        }
    }
    println("Trade direction is " + tradeDirection)
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
    } else if mode[0] == 's' || mode[0] == 'S' {
        if len(mode) < 2 {
            mode = "server"
        } else if mode[0:2] == "st" || mode[0:2] == "St" || mode[0:2] == "ST" || mode[0:2] == "sT" {
            mode = "stopMarket"
        } else {
            mode = "server"
        }
    } else if mode[0] == 'r' || mode[0] == 'R' {
        mode = "records"
    } else {
        if mode[0:2] == "ca" || mode[0:2] == "Ca" || mode[0:2] == "CA" || mode[0:2] == "cA" {
            mode = "cancel"
        } else if mode[0:2] == "cl" || mode[0:2] == "Cl" || mode[0:2] == "CL" || mode[0:2] == "cL" {
            mode = "close"
        } else {
            log.Fatal("Mode flag used with an unsuitable argument.")
        }
    }
}

func getTradingSymbol(asset string) string {
    return asset + "BTC"
}

func getMapKeys(m map[string]float64) []string {
    keys := make([]string, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

func getPrefix(str string) string {
    strLen := len(str)
    return str[:strLen-4]
}

func trimTrailingZeros(str string) string {
    return strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(str, "000"), "00"), "0")
}

//end getopt initialization
//invoke like ./main -dir=short -mult=1.0
/*
 * function main
 * params:
 ************************
 */
func main() {

    getopt.Parse()
    cleanFlagArguments()
    // printFlagArguments()
    println("made client")
    println(mode)
    marginAccount, _ := client.NewGetMarginAccountService().Do(context.Background())

    if mode == "market" {
        //line 630
        if len(pairsFlag) == 0 {
            marketOrders(client, stringToSide(tradeDirection), marginAccount.UserAssets)
        } else {
            marketOrders(client, stringToSide(tradeDirection), convertPairsToBinanceUserAssets(pairs))
        }
    } else if mode == "limit" {
        // limitOrders(client, stringToSide(tradeDirection), globalOffset, marginAccount.UserAssets)
        if len(pairsFlag) == 0 {
            limitOrders(client, stringToSide(tradeDirection), globalOffset, marginAccount.UserAssets)
        } else {
            println("we are just doing pairssss")
            limitOrders(client, stringToSide(tradeDirection), globalOffset, convertPairsToBinanceUserAssets(pairs))
        }
    } else if mode == "account" {
        account, _ := client.NewGetMarginAccountService().Do(context.Background())
        println(account)
    } else if mode == "cancel" {
        // cancelOrders(client, getOrders(client))

        // if floatFree == 0 && floatLocked == 0 { continue }

        floatOpenPositions := make(map[string]float64)
        for k := range positionPrecisions {
            floatOpenPositions[k] = 0.0
        }
        println("now going to cancel")

        //cancel the open spot orders
        // cancelSpotOrders(getOpenSpotOrders(floatOpenPositions))
        cancelMarginOrders(getOpenMarginOrders(floatOpenPositions))
    } else if mode == "close" {
        //getMarginAccount

        //getOpenMarginPositions
        openMarginPositions := getOpenMarginPositions(marginAccount.UserAssets)
        // println(openMarginPositions)
        //cancel the open margin orders
        cancelMarginOrders(getOpenMarginOrders(openMarginPositions))

        println("about to do closemarginpisitions")
        //close the open margin positions
        closeMarginPositions(openMarginPositions)
    } else if mode == "server" {
        setupServer()
    } else if mode == "records" {
        // retrieveRecords()
        _, r := client.NewListMarginLoansService().Asset(pairs[0]).Do(context.Background())
        fmt.Println(r)
    } else {
        log.Fatal("Utility not called with a suitable argument to the mode flag. Exiting without execution.")
    }
}

func getOpenSpotOrders(openPositions map[string]float64) []*binance.Order {
    openOrdersAcrossAllPairs := []*binance.Order{}
    for asset := range openPositions {
        openOrders, err := client.NewListOpenOrdersService().Symbol(asset).
            Do(context.Background())
        if err != nil {
            println(asset)
            fmt.Println(err)
            return nil
        }
        for _, order := range openOrders {
            // fmt.Println(order)
            openOrdersAcrossAllPairs = append(openOrdersAcrossAllPairs, order)
        }
    }
    return openOrdersAcrossAllPairs
}

func cancelSpotOrders(orders []*binance.Order) {
    for _, order := range orders {
        _, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}


func getLowestStopLoss(orders []*binance.Order) *binance.Order {
    var lowOrder *binance.Order
    for _, o := range orders {
        if o.Type == binance.OrderTypeStopLoss {
            if lowOrder == nil {
                lowOrder = o
            } else if o.StopPrice < lowOrder.StopPrice {
                lowOrder = o
            }   
        } else if o.Type == binance.OrderTypeStopLossLimit {
            if lowOrder == nil {
                lowOrder = o
            } else if o.StopPrice < lowOrder.StopPrice {
                lowOrder = o
            }   
        }
    }
    return lowOrder
}  

func getHighestStopLoss(orders []*binance.Order) *binance.Order{
    var highOrder *binance.Order 
    for _, o := range orders {
        if o.Type == binance.OrderTypeStopLoss {
            if highOrder == nil {
                highOrder = o
            } else if o.StopPrice > highOrder.StopPrice {
                highOrder = o
            }   
        } else if o.Type == binance.OrderTypeStopLossLimit {
            if highOrder == nil {
                highOrder = o
            } else if o.StopPrice > highOrder.StopPrice {
                highOrder = o
            }   
        }
    }
    return highOrder
} 

func stringToFloat(num string) float64 {
    floatBorrowed, errBorrowed := strconv.ParseFloat(num, 64)
    if errBorrowed != nil {
        log.Fatal("ERR INN STRING TO FLOAT")        
    }
    return floatBorrowed
}  

func cancelStop(stop *binance.Order) {
    _, cancelErr := client.NewCancelMarginOrderService().Symbol(stop.Symbol).OrderID(stop.OrderID).Do(context.Background())
    if cancelErr != nil {
        errorTokens := strings.Split(cancelErr.Error(), " ")
        //TODO: investigate library and structs and whether i am handling this wrong. Non-breaking
        if !stringsContains(errorTokens, "unmarshal") && !stringsContains(errorTokens, "CancelOrderResponse.orderId") { 
            // println("got an error canceling stop in closeMarginPositions for " + k)
        // } else {
            println("got an error canceling stop in cancelStop for " + stop.Symbol)
            log.Fatal(cancelErr)
        }
    }   
}

func closeMarginPositions(openPositions map[string]float64) {
        println("GOT THROUG EXCLUSIons")    
    for k, v := range openPositions {
        if pairsFlag != "" {
            if !Includes(pairs, k) {
                continue
            }
        }
        if k == "BTC" {
            continue
        }       

        if stringsContains(excludedPairs, k) {
            continue
        }

        if minimumPositionSizes[k] == 0 {
            continue
        }

        if math.Abs(v) < minimumPositionSizes[k] {
            continue
        }
        if v > 0 {
            println(getTradingSymbol(k))
            openOrders, listOrdersErr := client.NewListMarginOpenOrdersService().Symbol(getTradingSymbol(k)).Do(context.Background())
            println(len(openOrders))
            if listOrdersErr != nil {
                println("got an error listing orders in closeMarginPositions for " + k)
                log.Fatal(listOrdersErr)
            }
            lowStop := getLowestStopLoss(openOrders)
            if lowStop != nil {
                cancelStop(lowStop)
                marketOrder(client, binance.SideTypeSell, k, calculateOrderSizeFromPrecision(k, v*proportion))                
                //only put on a new stop loss if we are not closing the full position
                if proportion != 1.0 {
                    stopOrder(client, lowStop.Side, k, calculateOrderSizeFromPrecision(k,stringToFloat(lowStop.OrigQuantity) - stringToFloat(lowStop.ExecutedQuantity) - stringToFloat(calculateOrderSizeFromPrecision(k, v*proportion))), lowStop.Price, lowStop.Type)                    
                }
            } else {
                marketOrder(client, binance.SideTypeSell, k, calculateOrderSizeFromPrecision(k, v*proportion))                                
            }
        } else {
            openOrders, listOrdersErr := client.NewListMarginOpenOrdersService().Symbol(getTradingSymbol(k)).Do(context.Background())
            if listOrdersErr != nil {
                println("got an error listing orders in closeMarginPositions for " + k)
                log.Fatal(listOrdersErr)
            }
            highStop := getHighestStopLoss(openOrders)
            if highStop == nil {
                marketOrder(client, binance.SideTypeBuy, k, calculateOrderSizeFromPrecision(k, -1*v))                                
            } else {  
                cancelStop(highStop)         
                marketOrder(client, binance.SideTypeBuy, k, calculateOrderSizeFromPrecision(k, -1*v))                                
                //only put on a new stop loss if we are not closing the full position
                if proportion != 1.0 {
                    stopOrder(client, highStop.Side, k, calculateOrderSizeFromPrecision(k,stringToFloat(highStop.OrigQuantity) - stringToFloat(highStop.ExecutedQuantity) - stringToFloat(calculateOrderSizeFromPrecision(k, -1*v*proportion))), highStop.Price, highStop.Type)                    
                }                
            }
        }
    }
}

func getOpenMarginPositions(assets []binance.UserAsset) map[string]float64 {
    println("assets: ")
    openPositions := make(map[string]float64)
    for _, userAsset := range assets {
        // for i, userAsset := range assets {
        // if userAsset.Asset == "BTC" || userAsset.Asset == "USDT" {
        if userAsset.Asset == "BTC" || userAsset.Asset == "USDT" || userAsset.Asset == "BUSD" {
            continue
        }
        floatFree, errFree := strconv.ParseFloat(userAsset.Free, 64)
        if errFree != nil {
            println(userAsset.Asset)
            log.Fatal(errFree)
        }
        floatLocked, errLocked := strconv.ParseFloat(userAsset.Locked, 64)
        if errLocked != nil {
            println(userAsset.Asset)
            log.Fatal(errLocked)
        }
        floatBorrowed, errBorrowed := strconv.ParseFloat(userAsset.Borrowed, 64)
        if errBorrowed != nil {
            println(userAsset.Asset)
            log.Fatal(errBorrowed)
        }
        if floatFree == 0 && floatLocked == 0 && floatBorrowed == 0 {
            continue
        }
        openPositions[userAsset.Asset] = floatFree + floatLocked - floatBorrowed
    }
    return openPositions
}

func setupServer() {
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.HandleFunc("/", serveTemplate)
    // http.HandleFunc("/market_buy.json", marketBuyHandler)
    // http.HandleFunc("/market_sell.json", marketSellHandler)
    // http.HandleFunc("/limit_buy.json", limitBuyHandler)
    // http.HandleFunc("/limit_sell.json", limitSellHandler)
    // http.HandleFunc("/close.json", closeHandler)
    // http.HandleFunc("/cancel.json", cancelHandler)
    log.Println("Listening...")
    http.ListenAndServe(":8080", nil)
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
    lp := filepath.Join("templates", "layout.html")
    fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

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

// func marketBuyHandler(w http.ResponseWriter, r *http.Request) {
//     tradeDirection = "BUY"
//     marketOrders(client, stringToSide(tradeDirection))
//     // openDirection = "BUY"
//     fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
// }

// func marketSellHandler(w http.ResponseWriter, r *http.Request) {
//     tradeDirection = "SELL"
//     marketOrders(client, stringToSide(tradeDirection))
//     openDirection = "SELL"
//     fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
// }

// func limitBuyHandler(w http.ResponseWriter, r *http.Request) {
//     tradeDirection = "BUY"
//     // limitOrders()
//     limitOrders(client, stringToSide(tradeDirection), globalOffset)
//     openDirection = "BUY"
//     fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
// }

// func limitSellHandler(w http.ResponseWriter, r *http.Request) {
//     tradeDirection = "SELL"
//     // limitOrders()
//     limitOrders(client, stringToSide(tradeDirection), globalOffset)
//     openDirection = "SELL"
//     fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
// }

func cancelHandler(w http.ResponseWriter, r *http.Request) {
    // cancelOrders(client, getOrders(client))
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func closeHandler(w http.ResponseWriter, r *http.Request) {
    // closeOpenPositions(client)
    //TODO FINISH
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func cancelMarginOrders(orders []*binance.Order) {
    for _, order := range orders {
        //if we are only closing a portion of the position, then leave the stop losses open 
        if order.Type == binance.OrderTypeStopLoss || order.Type == binance.OrderTypeStopLossLimit {
            continue
        }
        _, err := client.NewCancelMarginOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}

func getOpenMarginOrders(openPositions map[string]float64) []*binance.Order {
    openOrdersAcrossAllPairs := []*binance.Order{}
    fmt.Println("map:", openPositions)
    for asset := range openPositions {
        println(asset)
        openOrders, err := client.NewListMarginOpenOrdersService().Symbol(getTradingSymbol(asset)).
            Do(context.Background())
        if err != nil {
            fmt.Println(err)
            return nil
        }
        for _, order := range openOrders {
            fmt.Println(order)
            openOrdersAcrossAllPairs = append(openOrdersAcrossAllPairs, order)
        }
    }
    return openOrdersAcrossAllPairs
}

/*
 * function getPrice
 * params: client
 ********
 */
func getPrice(client *binance.Client, asset string) []*binance.SymbolPrice {
    price, err := client.NewListPricesService().Do(context.Background())
    if err != nil {
        fmt.Println(err)
        return nil
    }
    //skip btc and eth prices
    return price
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
        return nil
    }
    for k, v := range prices {
        println(k)
        println(v)
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
 * function HelloServer
 * params: client
 ************************
 */
func HelloServer(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!<br><button><a href='/'>Buy</a></button><br><button><a href='/'>Sell</a></button><br><button><a href='/'>Close</a></button><br>", r.URL.Path[1:])
}

func round(x, unit float64) float64 {
    return math.Round(x/unit) * unit
}

/*
 * function limitOrders
 * params: client
 ************************
 */
func limitOrders(client *binance.Client, direction binance.SideType, offset float64, assets []binance.UserAsset) {
    // prices := getPrices(client)
    // for _, price := range prices {
    //     fmt.Println(price)
    // }
    // for k, v := range prices{
    //     println(k)
    //     println(v)
    // }
    for globalPairIndex, userAsset := range assets {
        println(userAsset.Asset)
        price := getPrice(client, userAsset.Asset)[0]
        println(price.Price)
        reversedIndex := len(assets) - globalPairIndex - 1
        println(direction)
        //bugg -- need to get the right price of the asset
        // println(prices[reversedIndex].Price)
        size := calculateOrderSizeFromPrecision(userAsset.Asset, minimumPositionSizes[userAsset.Asset])
        // println(calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, assets[reversedIndex].Asset))
        println(calculateOrderPriceFromOffset(price.Price, offset, direction, assets[reversedIndex].Asset))
        // limitOrder(client, direction, userAsset.Asset, offset, size, calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, assets[reversedIndex].Asset))
        limitOrder(client, direction, userAsset.Asset, offset, size, calculateOrderPriceFromOffset(price.Price, offset, direction, assets[reversedIndex].Asset))
    }
}

/*
 * function calculateOrderSizeFromPrecision
 * params: priceString string, offset float64, direction binance.SideType
 ************************
 */
func calculateOrderSizeFromPrecision(asset string, size float64) string {
    size = math.Floor(size*math.Pow10(positionPrecisions[asset])) / float64(math.Pow10(positionPrecisions[asset]))
    if positionPrecisions[asset] == 0 {
        return fmt.Sprintf("%d", int64(size))
    } else if positionPrecisions[asset] == 1 {
        return fmt.Sprintf("%.1f", size)
    } else if positionPrecisions[asset] == 2 {
        return fmt.Sprintf("%.2f", size)
    } else if positionPrecisions[asset] == 3 {
        return fmt.Sprintf("%.3f", size)
    } else if positionPrecisions[asset] == 4 {
        return fmt.Sprintf("%.4f", size)
    } else if positionPrecisions[asset] == 5 {
        return fmt.Sprintf("%.5f", size)
    } else if positionPrecisions[asset] == 6 {
        return fmt.Sprintf("%.6f", size)
    } else if positionPrecisions[asset] == 7 {
        return fmt.Sprintf("%.7f", size)
    } else {
        return fmt.Sprintf("%.8f", size)
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
    println(price)
    println("the offset is ")
    println(priceOffset)
    println(price - priceOffset)
    if pricePrecisions[asset] == 2 {
        return fmt.Sprintf("%.2f", price-priceOffset)
    } else if pricePrecisions[asset] == 3 {
        return fmt.Sprintf("%.3f", price-priceOffset)
    } else if pricePrecisions[asset] == 4 {
        return fmt.Sprintf("%.4f", price-priceOffset)
    } else {
        return fmt.Sprintf("%.5f", price-priceOffset)
    }
}

/*
 * function limitOrder
 * params: client
 ************************
 */
func limitOrder(client *binance.Client, direction binance.SideType, asset string, offset float64, size string, price string) {
    order, err := client.NewCreateMarginOrderService().Symbol(getTradingSymbol(asset)).
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
 * function marketOrders
 * params: client
 ************************
    //called on lin 230
*/
func marketOrders(client *binance.Client, direction binance.SideType, assets []binance.UserAsset) {
    //first -- get account and see what needs to be borrowed
    //only borrow if we are short selling
    if direction == binance.SideTypeSell {
        for _, userAsset := range assets {
            if userAsset.Asset == "BTC" || userAsset.Asset == "USDT" {
                continue
            }
            floatBorrowed, errBorrowed := strconv.ParseFloat(userAsset.Borrowed, 64)
            if errBorrowed != nil {
                println(userAsset.Asset)
                println("we got error!!!")
                log.Fatal(errBorrowed)
            }
            if floatBorrowed == 0 {
                res := client.NewMarginLoanService().Asset(userAsset.Asset).Amount(calculateOrderSizeFromPrecision(userAsset.Asset, minimumPositionSizes[userAsset.Asset]))
                client.NewMarginLoanService().Asset(userAsset.Asset).Amount(calculateOrderSizeFromPrecision(userAsset.Asset, minimumPositionSizes[userAsset.Asset]))
                fmt.Println(res)
                userAsset.Borrowed = calculateOrderSizeFromPrecision(userAsset.Asset, minimumPositionSizes[userAsset.Asset])
            }
        }
    }
    for _, asset := range pairs {
        marketOrder(client, direction, asset, calculateOrderSizeFromPrecision(asset, positionMultiplier*minimumPositionSizes[asset]))

        //if we need to borrow, go ahead and do it

        // res := client.NewMarginLoanService().Asset(userAsset.Asset).Amount(calculateOrderSizeFromPrecision(userAsset.Asset, minimumPositionSizes[userAsset.Asset]))
    }
}

/*
 * function marketOrder
 * params: client
 ************************
 * returns error (optionally)
 */
//TODO: integrate openPositions using mongodb
func marketOrder(client *binance.Client, direction binance.SideType, asset string, size string) error {
    marginOrder, marginOrderErr := client.NewCreateMarginOrderService().Symbol(getTradingSymbol(asset)).
        Side(direction).Type(binance.OrderTypeMarket).
        Quantity(size).SideEffectType(binance.SideEffectTypeAutoRepay).Do(context.Background())
    if marginOrderErr != nil {
        println(asset)
        fmt.Println(marginOrderErr)
    }
    //if we need to borrow, go ahead and do it

    // res := client.NewMarginLoanService().Asset(userAsset.Asset).Amount(calculateOrderSizeFromPrecision(userAsset.Asset, minimumPositionSizes[userAsset.Asset]))
    // openPositions[asset] -= positionMultiplier * positionSizes[asset]
    fmt.Println(marginOrder)
    return nil
}

/*
 * function marketOrder
 * params: client
 ************************
 * returns error (optionally)
 */
//TODO: integrate openPositions using mongodb
func stopOrder(client *binance.Client, direction binance.SideType, asset string, size string, price string, orderType binance.OrderType) error {
    stopOrder, stopErr := client.NewCreateMarginOrderService().Symbol(getTradingSymbol(asset)).
        Side(direction).Type(orderType).Quantity(size).TimeInForce(binance.TimeInForceTypeGTC).Price(price).StopPrice(price).
        SideEffectType(binance.SideEffectTypeAutoRepay).Do(context.Background())                
    if stopErr != nil {
        println("got an error making new stop in closeMarginPositions for " + asset)
        log.Fatal(stopErr)                            
    }
    println("Stop Order: ")
    fmt.Println(stopOrder)
    return nil
}
