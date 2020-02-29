package main

import (
    "fmt"
    "net/http"
    "github.com/ggarza5/go-binance-margin"
    "github.com/pborman/getopt/v2"
    "context"
    "log"
    "html/template"
    "os"
    "path/filepath"
    "math"
    "strconv"
    "strings"
)

//TODO: open positions
var (
    apiKey = "jIyd39L4YfD5CRvygwh5LY1IVilQ38NXY5RshUxKGwR1Sjj6ZGzynkxfK1p2jX0c"            
    secretKey = "3IbVAdTpwMN417BNbiwxc63NMpm0EZiBRbC7YFol4gbMytV4FxtfBfJ5dGkgq5Z2"
    openDirection = "SELL"
    tradeDirection = "BUY"
    globalOffset = 0.01
    positionMultiplier = 0.1

    globalOffsetFlag = "0.01"
    positionMultiplierFlag = "0.1"
    mode = "market"

    client = binance.NewClient(apiKey, secretKey)
    //these are the pairs that are not on bitmex, so i would onyl sohrt them on binance
    pairs = []string{"MATIC", "BNB", "LINK", "ATOM", "DASH", "ZEC", "ONT", "NEO", "XTZ", "ETC", "IOST", "IOTA", "XLM", "RVN", "BAT", "QTUM", "XMR", "VET"}
    //current
    accountSize = 0.1
    //TODO: change to calculate these minimumPositionSizes on the fly by pulling prices
    minimumPositionSizes = map[string]float64 {
        "MATIC": 50,
        "BNB": 0.05,
        "LINK": 1,
        "ATOM": 0.25,
        "DASH": 0.02,
        "ZEC": 0.02,
        "ONT": 1.4,
        "NEO": 0.08,
        "XTZ": 0.5,
        "ETC": 0.12,
        "IOST": 185,
        "IOTA": 4,
        "XLM": 15,
        "RVN": 40,
        "BAT": 5,
        "QTUM": 0.5,
        "XMR": 0.02,
        "VET": 160,
    }
    positionPrecisions = map[string]int {
        "BNB":2,        
        "LINK":0, 
        "ATOM":2,       
        "EOS":2,
        "DASH":3,               
        "ZEC":3,                     
        "ONT":2,
        "NEO":2,
        "LTC":3,    
        "MATIC":0,
        "XTZ":2,
        "ETC":2,
        "IOST":0,
        "IOTA":0,
        "XLM":0,    
        "RVN":0,
        "BAT":0,
        "QTUM":2,   
        "XMR":3,
        "VET":0,
    } 

    pricePrecisions = map[string]int {   
        "ZEC":6,              
        "LINK":8,
        "NEO":6,       
        "ETH":6,       
        "DASH":6,
        "TRX":8,        
        "LTC":6,      
        "MATIC":8,       
        "XTZ":7,
        "ETC":7,
        "IOST":8,
        "IOTA":8,
        "XLM":8,        
        "RVN":8,
        "BNB":7,
        "BAT":8,
        "QTUM":7,       
        "VET":8,
        "ATOM":7,      
        "XMR":6,     
    }
)

func contains(s []string, e string) bool {
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
}

func getAccount(client *binance.Client) *binance.Account {
    account, _  := client.NewGetAccountService().Do(context.Background())
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
}

//TODO:genericize cleaning of arguments using "reflect" package and interface pointers
func cleanFlagArguments() {
    if tradeDirection[0] == ' ' || tradeDirection[0] == '=' {
        tradeDirection = tradeDirection[1:]
    }
    if tradeDirection[0] == 's' || tradeDirection[0] == 'S' {
        tradeDirection = "SELL"
    } else if tradeDirection[0] == 'l' || tradeDirection[0] == 'L' {
        tradeDirection = "BUY"
    } else {
        log.Fatal("Direction flag used with an unsuitable argument.")        
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
    }  else if mode[0] == 's'|| mode[0] == 'S' {
        mode = "server"
    }  else if mode[0] == 'r'|| mode[0] == 'R' {
        mode = "records"
    } else {
        if mode[0:2] == "ca" || mode[0:2] == "Ca" || mode[0:2] == "CA" || mode[0:2] == "cA" {
            mode = "cancel"
        }  else if mode[0:2] == "cl" || mode[0:2] == "Cl" || mode[0:2] == "CL" || mode[0:2] == "cL" {
            mode = "close"
        }  else {
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

func getPrefix(str string) string{
    strLen := len(str)
    return str[:strLen-4]
}

func trimTrailingZeros(str string) string {
    return strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(str, "000"), "00"),"0")
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
        marketOrders(client, stringToSide(tradeDirection), marginAccount.UserAssets)
    } else if mode == "limit" {
        // limitOrders(client, stringToSide(tradeDirection), globalOffset)
    } else if mode == "account" {
        account, _  := client.NewGetMarginAccountService().Do(context.Background())
        println(account)        
    } else if mode == "cancel" {
        // cancelOrders(client, getOrders(client))

        // if floatFree == 0 && floatLocked == 0 { continue }            

        floatOpenPositions := make(map[string]float64)
        for k, _ := range positionPrecisions {
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

func closeMarginPositions(openPositions map[string]float64) {
    println("closign margin START")
    for k, v := range openPositions {
        // var closeMarginOrder *binance.Order
        // var closeOrderErr err
        if k == "BTC" { continue }
        if minimumPositionSizes[k] == 0 { continue }
        if math.Abs(v) < minimumPositionSizes[k] { continue }
        if v > 0 {
            // amt := fmt.Sprintf("%.8f", v) 
            // if k == "BNB" {
            //     amt = fmt.Sprintf("%d", int64(v))
            // }
            // if amt == "0" { continue }
            // println(amt)
            // println(getTradingSymbol(k))
            marginCloseOrder, marginOrderErr := client.NewCreateMarginOrderService().Symbol(getTradingSymbol(k)).
            Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
            Quantity(calculateOrderSizeFromPrecision(k, v)).SideEffectType(binance.SideEffectTypeAutoRepay).Do(context.Background())
            println(marginCloseOrder)
            if marginOrderErr != nil {
                println(k)
                fmt.Println(marginOrderErr)
            }
        } else {
            // amt := fmt.Sprintf("%.8f", -1*v) 
            // println(k)
            // if k == "BNB" {
            //     amt = fmt.Sprintf("%d", int64(-1*v))
            // }                       
            marginCloseOrder, marginOrderErr := client.NewCreateMarginOrderService().Symbol(getTradingSymbol(k)).
            Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).
            Quantity(calculateOrderSizeFromPrecision(k, -1*v)).SideEffectType(binance.SideEffectTypeAutoRepay).Do(context.Background())
            println(marginCloseOrder)
            if marginOrderErr != nil {
                println(k)
                fmt.Println(marginOrderErr)
            }
        }
    }
    println("closign margin END")    
}

func getOpenMarginPositions(assets []binance.UserAsset) map[string]float64 {
    println("assets: ")
    openPositions := make(map[string]float64)        
    for _, userAsset := range assets {
    // for i, userAsset := range assets {
        if userAsset.Asset == "BTC" || userAsset.Asset == "USDT" { continue }
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
        if floatFree == 0 && floatLocked == 0 && floatBorrowed == 0 { continue }
        openPositions[userAsset.Asset] = floatFree + floatLocked - floatBorrowed
        // localPairs = append(localPairs, getTradingSymbol(userAsset.Asset))
        // println(i)
        // println(userAsset.Asset)
        // println(userAsset.Borrowed)
        // println(userAsset.Free)
        // println(userAsset.Interest)
        // println(userAsset.Locked)
        // println(userAsset.NetAsset)
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

func cancelMarginOrders(orders[]*binance.Order) {
    for _, order := range orders {
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
// func limitOrders(client *binance.Client, direction binance.SideType, offset float64) {
//     prices := getPrices(client)
//     for globalPairIndex, _ := range pairs {
//         reversedIndex := len(pairs) - globalPairIndex - 1
//         size := calculateOrderSizeFromPrecision(positionMultiplier * positionSizes[pairs[reversedIndex]], pairs[reversedIndex])
//         limitOrder(client, direction, pairs[reversedIndex], offset, size, calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, pairs[reversedIndex]))
//     }
// }

/*
 * function calculateOrderSizeFromPrecision
 * params: priceString string, offset float64, direction binance.SideType
 ************************
 */
func calculateOrderSizeFromPrecision(asset string, size float64) string {
    size = math.Floor(size * math.Pow10(positionPrecisions[asset]))/float64(math.Pow10(positionPrecisions[asset]))
    if positionPrecisions[asset] == 0 {
        return fmt.Sprintf("%d", int64(size))
    } else if (positionPrecisions[asset] == 1) { 
        return fmt.Sprintf("%.1f", size)
    } else if (positionPrecisions[asset] == 2) { 
        return fmt.Sprintf("%.2f", size)
    } else if (positionPrecisions[asset] == 3) {
        return fmt.Sprintf("%.3f", size)
    } else if (positionPrecisions[asset] == 4) { 
        return fmt.Sprintf("%.4f", size)
    } else if (positionPrecisions[asset] == 5) { 
        return fmt.Sprintf("%.5f", size)
    } else if (positionPrecisions[asset] == 6) { 
        return fmt.Sprintf("%.6f", size)
    } else if (positionPrecisions[asset] == 7) { 
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
// func calculateOrderPriceFromOffset(priceString string, offset float64, direction binance.SideType, asset string) string {
//     price, err := strconv.ParseFloat(priceString, 64)
//     if err != nil {
//         fmt.Println(err)
//         return ""
//     }
//     var priceOffset float64
//     if (direction == binance.SideTypeBuy) { 
//         priceOffset = price * offset 
//     } else { 
//         priceOffset = price * offset * -1 
//     }    
//     if (pricePrecisions[asset] == 2) {         
//         return fmt.Sprintf("%.2f", price - priceOffset)
//     } else if (pricePrecisions[asset] == 3) {        
//         return fmt.Sprintf("%.3f", price - priceOffset)
//     } else if (pricePrecisions[asset] == 4) {         
//         return fmt.Sprintf("%.4f", price - priceOffset)
//     } else {         
//         return fmt.Sprintf("%.5f", price - priceOffset) 
//     }            
// }

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
 * function marketOrders
 * params: client
 ************************
 */
func marketOrders(client *binance.Client, direction binance.SideType, assets []binance.UserAsset) {
    //called on lin 230
    //first -- get account and see what needs to be borrowed
    //only borrow if we are short selling
    if direction == binance.SideTypeSell {
        for _, userAsset := range assets {
            if userAsset.Asset == "BTC" || userAsset.Asset == "USDT" { continue }   
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
        marketOrder(client, direction, asset, calculateOrderSizeFromPrecision(asset, positionMultiplier * minimumPositionSizes[asset]))
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
    // openPositions[asset] -= positionMultiplier * positionSizes[asset]        
    fmt.Println(marginOrder)
    return nil
}
