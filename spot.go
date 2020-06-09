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

var (
    apiKey = "jIyd39L4YfD5CRvygwh5LY1IVilQ38NXY5RshUxKGwR1Sjj6ZGzynkxfK1p2jX0c"
    secretKey = "3IbVAdTpwMN417BNbiwxc63NMpm0EZiBRbC7YFol4gbMytV4FxtfBfJ5dGkgq5Z2"
    openDirection = "SELL"
    tradeDirection = ""
    globalOffset = 0.01
    positionMultiplier = 1.0

    globalOffsetFlag = "0.01"
    positionMultiplierFlag = "1.0"
    mode = "market"

    client = binance.NewClient(apiKey, secretKey)
    priceSlice = []*binance.SymbolPrice{}
    prices = make(map[string]string)

    pairs = []string{}
    pairsFlag = ""
    entriesFlag = ""
    entries = []string{"0.0001"}

    slFlag = ""
    sl = "0"
    tpFlag = ""
    tps = []string{"1000000"}

    minimumOrderSize = 0.0001
    positionSize = 0.002

    positionPrecisions = map[string]int {
        "ETH":3,
        "BNB":2,        
        "LINK":0, 
        "XRP":0,                       
        "ALGO":0,        
        "ATOM":2,
        "XEM":0,                
        "FET":0,
        "IOTX":0,
        "PHB":0,
        "CELR":0,
        "CHZ":0,
        "DLT":0,
        "EOS":2,
        "GRS":0,
        "KEY":0,
        "MITH":0,
        "ARN":0,
        "WRX":0,
        "ZEC":3,
        "BTG":2,
        "DREP":0,
        "GVT":2,
        "REN":0,
        "SC":0,
        "DOCK":0,
        "BQX":0,
        "FTT":2,
        "KMD":2,
        "NAV":0,
        "ANKR":0,
        "BCD":0,
        "BRD":0,
        "GNT":0,
        "MCO":2,
        "MTL":0,
        "REP":3,
        "STORJ":0,
        "BAND":0,
        "DUSK":0,
        "LSK":2,
        "MFT":0,
        "POLY":0,
        "THETA":0,
        "DNT":0,
        "DOGE":0,
        "GTO":0,
        "OMG":2,
        "OST":0,
        "NAS":2,
        "RCN":0,
        "TOMO":0,
        "ZIL":0,
        "LUN":2,
        "CDT":0,
        "NXS":0,
        "POE":0,
        "SNM":0,
        "SYS":0,
        "WPR":0,
        "XVG":0,
        "ARDR":0,
        "YOYO":0,
        "BNT":0,
        "DGD":3,
        "ENG":0,
        "KNC":0,
        "QSP":0,
        "APPC":0,
        "VITE":0,
        "PPT":0,
        "ARPA":0,
        "EVX":0,
        "KAVA":0,
        "ONT":2,
        "WABI":0,
        "WTC":2,
        "AION":0,
        "GAS":2,
        "NEO":2,
        "OGN":0,
        "ONE":0,
        "SKY":0,
        "CVC":0,
        "GO":0,
        "STX":0,
        "TROY":0,
        "AE":0,
        "LRC":0,
        "LTO":0,
        "NULS":0,
        "PIVX":0,
        "BLZ":0,
        "FUEL":0,
        "MDA":0,
        "VIA":0,
        "WAN":0,
        "DCR":3,
        "DASH":3,
        "EDO":0,
        "FUN":0,
        "HOT":0,
        "ICX":0,
        "INS":0,
        "MANA":0,
        "BTS":0,
        "STEEM":0,
        "TNT":0,
        "TRX":0,
        "VIBE":0,
        "WAVES":2,
        "OAX":0,
        "LOOM":0,
        "LTC":3,
        "STORM":0,
        "XZC":2,
        "ELF":0,
        "ENJ":0,
        "MATIC":0,
        "NKN":0,
        "PERL":0,
        "RLC":0,
        "STRAT":0,
        "VIB":0,
        "ARK":0,
        "XTZ":2,
        "DATA":0,
        "ETC":2,
        "GXS":0,
        "IOST":0,
        "IOTA":0,
        "XLM":0,
        "CND":0,
        "POA":0,
        "RVN":0,
        "LEND":0,
        "ZEN":2,
        "BCH":3,
        "BAT":0,
        "QKC":0,
        "TFUEL":0,
        "ADA":0,
        "COS":0,
        "MTH":0,
        "QTUM":2,
        "CMT":0,
        "AMB":0,
        "FTM":0,
        "SNGLS":0,
        "SNT":0,
        "TNB":0,
        "VET":0,
        "ADX":0,
        "BCPT":0,
        "BEAM":2,
        "POWR":0,
        "AGI":0,
        "ERD":0,
        "MBL":0,
        "ONG":0,
        "XMR":3,
        "COCOS":0,
        "HC":2,
        "RDN":0,
        "ZRX":0,
        "QLC":0,
        "CTXC":0,
        "HBAR":0,
        "NCASH":0,
        "NEBL":0,
        "REQ":0,
        "TCT":0,
        "AST":0,
        "NANO":2,
        "SOL":0,
    } 

    pricePrecisions = map[string]int {
        "FET":8,
        "IOTX":8,
        "PHB":8,
        "XEM":8,
        "CELR":8,
        "CHZ":8,
        "DLT":8,
        "EOS":7,
        "GRS":8,
        "KEY":8,
        "MITH":8,
        "ARN":8,
        "WRX":8,
        "ZEC":6,
        "BTG":6,
        "DREP":8,
        "GVT":7,
        "REN":8,
        "SC":8,
        "DOCK":8,
        "BQX":8,
        "FTT":7,
        "KMD":7,
        "NAV":8,
        "ANKR":8,
        "BCD":8,
        "BRD":8,
        "GNT":8,
        "MCO":7,
        "MTL":8,
        "REP":6,
        "STORJ":8,
        "BAND":8,
        "DUSK":8,
        "LSK":7,
        "MFT":8,
        "POLY":8,
        "THETA":8,
        "DNT":8,
        "DOGE":8,
        "GTO":8,
        "OMG":7,
        "OST":8,
        "ALGO":8,
        "NAS":7,
        "RCN":8,
        "TOMO":8,
        "ZIL":8,
        "LUN":7,
        "CDT":8,
        "NXS":8,
        "POE":8,
        "SNM":8,
        "SYS":8,
        "WPR":8,
        "XVG":8,
        "ARDR":8,
        "YOYO":8,
        "BNT":8,
        "DGD":6,
        "ENG":8,
        "KNC":8,
        "QSP":8,
        "APPC":8,
        "VITE":8,
        "PPT":8,
        "ARPA":8,
        "EVX":8,
        "KAVA":8,
        "ONT":7,
        "WABI":8,
        "WTC":7,
        "AION":8,
        "GAS":7,
        "LINK":8,
        "NEO":6,
        "OGN":8,
        "ONE":8,
        "SKY":8,
        "CVC":8,
        "ETH":6,
        "GO":8,
        "STX":8,
        "TROY":8,
        "AE":8,
        "LRC":8,
        "LTO":8,
        "NULS":8,
        "PIVX":8,
        "BLZ":8,
        "FUEL":8,
        "MDA":8,
        "VIA":8,
        "WAN":8,
        "DCR":6,
        "DASH":6,
        "EDO":8,
        "FUN":8,
        "HOT":8,
        "ICX":8,
        "INS":8,
        "MANA":8,
        "BTS":8,
        "STEEM":8,
        "TNT":8,
        "TRX":8,
        "VIBE":8,
        "WAVES":7,
        "OAX":8,
        "LOOM":8,
        "LTC":6,
        "STORM":8,
        "XZC":7,
        "ELF":8,
        "ENJ":8,
        "MATIC":8,
        "NKN":8,
        "PERL":8,
        "RLC":8,
        "STRAT":8,
        "VIB":8,
        "ARK":8,
        "XTZ":7,
        "DATA":8,
        "ETC":7,
        "GXS":8,
        "IOST":8,
        "IOTA":8,
        "XLM":8,
        "CND":8,
        "POA":8,
        "RVN":8,
        "BNB":7,
        "LEND":8,
        "ZEN":7,
        "BCH":6,
        "BAT":8,
        "QKC":8,
        "TFUEL":8,
        "ADA":8,
        "COS":8,
        "MTH":8,
        "QTUM":7,
        "CMT":8,
        "AMB":8,
        "FTM":8,
        "SNGLS":8,
        "SNT":8,
        "TNB":8,
        "VET":8,
        "ADX":8,
        "ATOM":7,
        "BCPT":8,
        "BEAM":7,
        "POWR":8,
        "AGI":8,
        "ERD":8,
        "MBL":8,
        "ONG":8,
        "XMR":6,
        "COCOS":8,
        "HC":7,
        "RDN":8,
        "ZRX":8,
        "QLC":8,
        "CTXC":8,
        "HBAR":8,
        "NCASH":8,
        "NEBL":8,
        "REQ":8,
        "TCT":8,
        "XRP":8,
        "AST":8,
        "NANO":7,
        "SOL":8,
    }
    positionSizes = map[string]float64{
        "ICX":1,
        "MATIC":1,
    }

)
//Done:
//Entries
//Todo:
//Exits, SL
//set global openPositions vriable and manage it
//If we don't have enough btc in spot account or margin account, make a MarginTransferService and transfer it, and if we only have USDT, then trade it for btc 

//2 options for SL
//Set limit orders to enter, then query those orders every minute or 5 minutes, then once they are filled, place stop orders
//enter with market orders and place the stop loss right away

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func isNonTradingCoin(asset string) bool {
    nonTradingSpotCoins := []string{"ETF", "BCX", "SBTC"}
    return contains(nonTradingSpotCoins, asset)
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
    getopt.FlagLong(&entriesFlag, "entries", 'e', "entries").SetOptional()    
    getopt.FlagLong(&slFlag, "stop loss", 's', "stop loss").SetOptional()  
    getopt.FlagLong(&tpFlag, "take profit", 't', "take profit").SetOptional()        
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
    println(pairsFlag)
    println(entriesFlag)
}

//TODO:genericize cleaning of arguments using "reflect" package and interface pointers
func cleanFlagArguments() {
    if len(pairsFlag) != 0 {
        if pairsFlag[0] == ' ' || pairsFlag[0] == '=' {
            pairsFlag = pairsFlag[1:]
        }        
        pairs = strings.Split(pairsFlag, ",")
        for _, p := range pairs {
            println(p)
        }
    }   
    if len(entriesFlag) != 0 {
        if entriesFlag[0] == ' ' || entriesFlag[0] == '=' {
            entriesFlag = entriesFlag[1:]
        }        
        entries = strings.Split(entriesFlag, ",")
    }  
    if len(slFlag) != 0 {
        if slFlag[0] == ' ' || slFlag[0] == '=' {
            slFlag = slFlag[1:]
        }        
    }  
    if len(tpFlag) != 0 {
        if tpFlag[0] == ' ' || tpFlag[0] == '=' {
            tpFlag = tpFlag[1:]
        }  
        tps = strings.Split(tpFlag, ",")              
    }                    
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

    globalOffset = parseFloatHandleErr(globalOffsetFlag)
    if positionMultiplierFlag[0] == ' ' || positionMultiplierFlag[0] == '=' {
        positionMultiplierFlag = positionMultiplierFlag[1:]
    }

    positionMultiplier = parseFloatHandleErr(positionMultiplierFlag)
    if mode[0] == ' ' || mode[0] == '=' {
        mode = mode[1:]
    }
    if mode[0] == 'm' || mode[0] == 'M' {
        mode = "market"
    } else if mode[0] == 'l'|| mode[0] == 'L' {
        mode = "limit"
    } else if mode[0] == 's'|| mode[0] == 'S' {
        if len(mode) < 2 {
            mode = "server"
        } else if (mode[0:2] == "st" || mode[0:2] == "St" || mode[0:2] == "ST" || mode[0:2] == "sT") {
            mode = "stopMarket"
        } else {
            mode = "server"
        }
    } else {
        // if mode[0:2] == "ca" || mode[0:2] == "Ca" || mode[0:2] == "CA" || mode[0:2] == "cA" {
        if strings.ToLower(mode[0:2]) == "ca" {
            mode = "cancel"
            println("we set mode to cancel")
        }  else if mode[0:2] == "cl" || mode[0:2] == "Cl" || mode[0:2] == "CL" || mode[0:2] == "cL" {
            mode = "close"
        }  else {
            log.Fatal("Mode flag used with an unsuitable argument.")
        }    
    }    
    // if tradeDirection[0] == ' ' || tradeDirection[0] == '=' {
    //     tradeDirection = tradeDirection[1:]
    // }

    // println("mode flag is " + mode[0:2])
    // if mode[0] == ' ' || mode[0] == '=' {
    //     mode = mode[1:]
    // }
    // if mode[0] == 'm' || mode[0] == 'M' {
    //     mode = "market"
    // } else if mode[0] == 'l'|| mode[0] == 'L' {
    //     mode = "limit"
    // }  else if mode[0:2] == "ca" || mode[0:2] == "Ca" || mode[0:2] == "CA" || mode[0:2] == "cA" {
    //     mode = "cancel"
    // }  else if mode[0:2] == "cl" || mode[0:2] == "Cl" || mode[0:2] == "CL" || mode[0:2] == "cL" {
    //     mode = "close"
    // } else if mode[0] == 's'|| mode[0] == 'S' {
    //     mode = "server"
    // } else {
    //     log.Fatal("Mode flag used with an unsuitable argument.")
    // }
    // println(mode)
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

func fillPriceMapFromTickers() {
    for _, symbolTicker := range priceSlice {
        prices[symbolTicker.Symbol] = symbolTicker.Price
    }
}

func Map(vs []string, f func(string) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}

func Filter(vs []*binance.SymbolPrice, f func(*binance.SymbolPrice) bool) []*binance.SymbolPrice {
    vsf := make([]*binance.SymbolPrice, 0)
    for _, v := range vs {
        if f(v) {
            vsf = append(vsf, v)
        }
    }
    return vsf
}

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

func isSymbolABTCTicker(s *binance.SymbolPrice) bool {
    index := strings.Index(s.Symbol, "BTC")
    if index != 0 && index != -1 {
        return true
    } else {
        return false
    }
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
    // println("we madddd")
    // return
    // printFlagArguments()
    println("Starting with mode - " + mode)

    account, _  := client.NewGetAccountService().Do(context.Background())  

    if len(pairs) >= 1 {
        priceSlice, _ = client.NewListPricesService().Symbol(strings.Join(Map(pairs, getTradingSymbol), ",")).Do(context.Background())
    // } else if len(pairs) == 1 {
    //     // returnedPrice, _ := client.NewListPricesService().Symbol(strings.Join(pairs, ",")).Do(context.Background())
    //     // prices = append(prices, returnedPrice)
    //     // priceSlice, _ = client.NewListPricesService().Symbol(getTradingSymbol(pairs[0])).Do(context.Background())        
    //     priceSlice, _ = client.NewListPricesService().Symbol(pairs[0]).Do(context.Background())        
    } else {
        priceSlice, _ = client.NewListPricesService().Do(context.Background())
        // fmt.Println(priceSlice)
    }
    priceSlice = Filter(priceSlice, isSymbolABTCTicker)
    fillPriceMapFromTickers()

    openSpotPositions := getOpenSpotPositions(account.Balances)      
    // fmt.Print(openSpotPositions)
    if mode == "market" {
        //if we don't have 0.006 BTC, just return. That is currently 2% of account size
        if openSpotPositions["BTC"] < 0.004 {
            println("We did not have enough BTC to enter a new position.")
            return
        }           
        marketOrders(client, stringToSide(tradeDirection))
        if slFlag != "" {
        // set OCO stop/TP orders
            for _, tp := range tps {
                highEntry := entries[1]
                numCoins := calculateNumberOfCoinsToBuy(highEntry)
                size := calculateOrderSizeFromPrecision(pairs[0], numCoins/2) 
                println(size)
                btcPrice := satsToBitcoin(tp)
                println(btcPrice)
                o, err := client.NewCreateOCOService().Symbol(getTradingSymbol(pairs[0])).Side(binance.SideTypeSell).StopPrice(fmt.Sprintf("%.8f", satsToBitcoin(slFlag))).
                StopLimitPrice(fmt.Sprintf("%.8f", satsToBitcoin(slFlag))).Price(fmt.Sprintf("%.8f", btcPrice)).Quantity(size).StopLimitTimeInForce("GTC").Do(context.Background())
                if err!= nil {
                    fmt.Println(err)
                }
                fmt.Println(o)
            }
        }
    } else if mode == "limit" {
        //if we don't have 0.006 BTC, just return. That is currently 2% of account size
        if openSpotPositions["BTC"] < 0.004 {
            println("We did not have enough BTC to enter a new position.")
            return
        }        
        limitOrders(client, stringToSide(tradeDirection), globalOffset)
    } else if mode == "account" {
        account, _  := client.NewGetAccountService().Do(context.Background())
        println(account)        
    } else if mode == "cancel" {
        //getSpotAccount
        
        //TODO: Convert to floatCopy that just has all the trading pairs
        //getOpenSpotPositions      
        openSpotPositions := getOpenSpotPositions(account.Balances)

        //cancel the open spot orders            
        cancelSpotOrders(getOpenSpotOrders(openSpotPositions))     
    } else if mode == "close" {
        account, _  := client.NewGetAccountService().Do(context.Background())
        // fmt.Println(account.Balances)
        //getOpenSpotPositions      
        openSpotPositions := getOpenSpotPositions(account.Balances)

        //cancel the open spot orders            
        cancelSpotOrders(getOpenSpotOrders(openSpotPositions))

        //close the open spot positions entirely
        closeSpotPositions(openSpotPositions)
    } else if mode == "server" {
        setupServer()
    } else {
        log.Fatal("Utility not called with a suitable argument to the mode flag. Exiting without execution.")
    }
}

//TODO: Get list of coins NOT to close from file provided at execution time
func getOpenSpotPositions(balances []binance.Balance) map[string]float64 {
    openSpotPositions := make(map[string]float64)  
    //get the pairs with open spot positions
    for _, balance := range balances {
        if isNonTradingCoin(balance.Asset) { continue }
        // if balance.Asset == "BTC" { continue }     
        if balance.Asset == "USDT" { continue }     
        if balance.Asset == "VTHO" { continue }

        if len(pairs) > 0 && !Includes(pairs, balance.Asset) && balance.Asset != "BTC" {
            continue
        }

        floatFree := parseFloatHandleErr(balance.Free)

        if balance.Asset == "BTC" {
            openSpotPositions[balance.Asset] = floatFree
            continue
        }

        floatLocked := parseFloatHandleErr(balance.Locked)          
        if (floatFree == 0 && floatLocked == 0) { continue }


        priceFloat := parseFloatHandleErr(prices[getTradingSymbol(balance.Asset)])
        if (floatFree + floatLocked) * priceFloat < minimumOrderSize {
            continue
        }
        openSpotPositions[balance.Asset] = floatFree + floatLocked
    }
    return openSpotPositions
}

func closeSpotPositions(openPositions map[string]float64) {
    if len(pairs) > 0 {
        for _, p := range pairs {
            println(p)
            tradingSymbol := getTradingSymbol(p)
            amt := calculateOrderSizeFromPrecision(tradingSymbol, openPositions[p])
            if amt == "0.0" { continue }

            println("now hee")
            // numCoins := calculateNumberOfCoinsToBuy(highEntry)
            // size := calculateOrderSizeFromPrecision(asset, numCoins*2) 
            println(calculateOrderSizeFromPrecision(p, openPositions[p]))
            spotCloseOrder, spotOrderErr := client.NewCreateOrderService().Symbol(tradingSymbol).
            Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
            Quantity(calculateOrderSizeFromPrecision(p, openPositions[p])).Do(context.Background())
            println(spotCloseOrder)
            if spotOrderErr != nil {
                fmt.Println(spotOrderErr)
            }        
        }       
    } else {
        for k, v := range openPositions {
            if k == "BTC" { continue }     
            println(getTradingSymbol(k))
            amt := calculateOrderSizeFromPrecision(k, v)
            if amt == "0.0" { continue }

            spotCloseOrder, spotOrderErr := client.NewCreateOrderService().Symbol(getTradingSymbol(k)).
            Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
            Quantity(calculateOrderSizeFromPrecision(k, v*positionMultiplier)).Do(context.Background())
            println(spotCloseOrder)
            if spotOrderErr != nil {
                fmt.Println(spotOrderErr)
            }
        }
    }
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

/***
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
***/

func cancelHandler(w http.ResponseWriter, r *http.Request) {
    // cancelOrders(client, getOrders(client))    
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func closeHandler(w http.ResponseWriter, r *http.Request) {
    // closeOpenPositions(client)
    //TODO FINISH
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func cancelSpotOrders(orders []*binance.Order) {
    // for _, pair := range pairs {
    //     _, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
    //     if err != nil {
    //         fmt.Println(err)
    //         return
    //     }        
    // }
    for _, order := range orders {      
        println(order.Symbol)
        _, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
        if err != nil {
            fmt.Println(err)
            return
        }
    }
}
   
func getOpenSpotOrders(openPositions map[string]float64) []*binance.Order {
    openOrdersAcrossAllPairs := []*binance.Order{}
    if len(pairs) > 0 {
        for _, asset := range pairs {
            openOrders, err := client.NewListOpenOrdersService().Symbol(getTradingSymbol(asset)).Do(context.Background())
            if err != nil {
                fmt.Println(err)
                return nil
            }
            for _, order := range openOrders {
                fmt.Println(order)            
                openOrdersAcrossAllPairs = append(openOrdersAcrossAllPairs, order)        
            }
        }        
    } else {
        for asset := range openPositions {
            println("canceling for " + asset)
            if asset == "BTC" {
                continue
            }
            openOrders, err := client.NewListOpenOrdersService().Symbol(getTradingSymbol(asset)).Do(context.Background())
            if err != nil {
                fmt.Println(err)
                return nil
            }
            for _, order := range openOrders {
                fmt.Println(order)            
                openOrdersAcrossAllPairs = append(openOrdersAcrossAllPairs, order)        
            }
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

func satsToBitcoin(u string) float64 {
    ii, _ := strconv.Atoi(u)
    return float64(ii) / math.Pow10(int(8))
}

func btcToSats(u string) int64 {
    ii, _ := strconv.ParseFloat(u,64)
    return int64(float64(ii) * math.Pow10(int(8)))
}


func ifOrderSizeMeetsMinimum(numCoins float64, entry string) bool {
    entryNum, _ := strconv.ParseFloat(entry, 64)
    if numCoins * entryNum >= minimumOrderSize {
        return true
    } else {
        return false
    }
}

/*
 * function limitOrders
 * params: client
 ************************
 */
func limitOrders(client *binance.Client, direction binance.SideType, offset float64) {
    for globalPairIndex, _ := range pairs {
        reversedIndex := len(pairs) - globalPairIndex - 1
        for _, entry := range entries {
            btcPrice := satsToBitcoin(entry)
            numCoins := calculateNumberOfCoinsToBuy(entry)/2
            if !ifOrderSizeMeetsMinimum(numCoins, entry) {
                continue
            }
            size := calculateOrderSizeFromPrecision(pairs[reversedIndex], calculateNumberOfCoinsToBuy(entry)/2)
            limitOrder(client, direction, getTradingSymbol(pairs[reversedIndex]), offset, size, fmt.Sprintf("%.8f", btcPrice))//calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, pairs[reversedIndex]))
        }
    }
}

// /*
//  * function stopMarketOrders
//  * params: client
//  ************************
//  */
// func stopMarketOrders(client *binance.Client, direction binance.SideType, offset float64, isReduceOnly binance.OrderReduceOnly, pairs ...string) {
//     prices := getPrices(client)  
//     for globalPairIndex, _ := range pairs {
//         reversedIndex := len(pairs) - globalPairIndex - 1        
//         size := calculateOrderSizeFromPrecision(pairs[reversedIndex], calculateNumberOfCoinsToBuy)        
//         stopMarketOrder(client, getOppositeDirection(direction), pairs[reversedIndex], size, calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, pairs[reversedIndex]), isReduceOnly)
//     }
// }

// /*
//  * function stopMarketOrder
//  * params: client
//  ************************
//  *///error-returning
// func stopMarketOrder(client *binance.Client, direction binance.SideType, asset string, size string, price string, isReduceOnly binance.OrderReduceOnly) error {
//     order, err := client.NewCreateOrderService().Symbol(asset).
//         Side(direction).Type(binance.OrderTypeStopLoss).
//         Quantity(size).ReduceOnly(isReduceOnly).
//         StopPrice(price).Do(context.Background())
//     if err != nil {
//         fmt.Println(err)
//         return err
//     }
//     // openPositions[asset] -= positionMultiplier * positionSizes[asset]        
//     fmt.Println(order)
//     return nil
// }

/*
 * function limitOrder
 * params: client
 ************************
 */
// func limitOrder(client *binance.Client, direction binance.SideType, asset string, offset float64, size string, price string) {
//     println("calling limit order")
//     order, err := client.NewCreateOrderService().Symbol(asset).
//             Side(direction).Type(binance.OrderTypeLimit).
//             TimeInForce(binance.TimeInForceTypeGTC).Quantity(size).
//             Price(price).Do(context.Background())
//     if err != nil {
//         fmt.Println(err)
//         return
//     }
//     fmt.Println(order)
// }

/*
 * function calculateOrderSizeFromPrecision
 * params: priceString string, offset float64, direction binance.SideType
 ************************
 */
func calculateOrderSizeFromPrecision(asset string, size float64) string {
    size = math.Floor(size * math.Pow10(positionPrecisions[asset]))/float64(math.Pow10(positionPrecisions[asset]))
    if positionPrecisions[asset] == 0 {
        return fmt.Sprintf("%d", int64(positionMultiplier*size))
    } else if (positionPrecisions[asset] == 1) { 
        return fmt.Sprintf("%.1f", positionMultiplier*size)
    } else if (positionPrecisions[asset] == 2) { 
        return fmt.Sprintf("%.2f", positionMultiplier*size)
    } else if (positionPrecisions[asset] == 3) {
        return fmt.Sprintf("%.3f", positionMultiplier*size)
    } else if (positionPrecisions[asset] == 4) { 
        return fmt.Sprintf("%.4f", positionMultiplier*size)
    } else if (positionPrecisions[asset] == 5) { 
        return fmt.Sprintf("%.5f", positionMultiplier*size)
    } else if (positionPrecisions[asset] == 6) { 
        return fmt.Sprintf("%.6f", positionMultiplier*size)
    } else if (positionPrecisions[asset] == 7) { 
        return fmt.Sprintf("%.7f", positionMultiplier*size)
    } else { 
        return fmt.Sprintf("%.8f", positionMultiplier*size) 
    }
}

/*
 * function calculateOrderSizeFromPrecision
 * params: priceString string, offset float64, direction binance.SideType
 ************************
 */
func calculateOrderPriceFromPrecision(asset string, price float64) string {
    price = math.Floor(price * math.Pow10(pricePrecisions[asset]))/float64(math.Pow10(pricePrecisions[asset]))
    if positionPrecisions[asset] == 0 {
        return fmt.Sprintf("%d", int64(price))
    } else if (positionPrecisions[asset] == 1) { 
        return fmt.Sprintf("%.1f", price)
    } else if (positionPrecisions[asset] == 2) { 
        return fmt.Sprintf("%.2f", price)
    } else if (positionPrecisions[asset] == 3) {
        return fmt.Sprintf("%.3f", price)
    } else if (positionPrecisions[asset] == 4) { 
        return fmt.Sprintf("%.4f", price)
    } else if (positionPrecisions[asset] == 5) { 
        return fmt.Sprintf("%.5f", price)
    } else if (positionPrecisions[asset] == 6) { 
        return fmt.Sprintf("%.6f", price)
    } else if (positionPrecisions[asset] == 7) { 
        return fmt.Sprintf("%.7f", price)
    } else { 
        return fmt.Sprintf("%.8f", price) 
    }
}

/*
 * function parseFloatHandleErr
 * params: floatString string
 * returns: float64
 ************************
 */
func parseFloatHandleErr(floatString string) float64 {
    ff, e := strconv.ParseFloat(floatString, 64)
    if e != nil {
        log.Fatal("Had an error when parsing float out of the following string: " + floatString)
    }
    return ff
}

//get the 
func calculateNumberOfCoinsToBuy(price string) float64 {
    ff := parseFloatHandleErr(price)
    numCoins := positionSize/ff
    // println(numCoins* math.Pow10(int(8)) )   
    return numCoins * math.Pow10(int(8))
    // if numCoins % 2 == 0 {
    //     return numCoins
    // } else {
    //     return numCoins - 1
    // }
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
    println("calling limit order")
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

// for globalPairIndex, _ := range pairs {
//         reversedIndex := len(pairs) - globalPairIndex - 1
//         println(reversedIndex)
//         // size := calculateOrderSizeFromPrecision(pairs[reversedIndex], positionMultiplier * positionSizes[pairs[reversedIndex]])
//         println(entries)
//         for _, entry := range entries {
//             println("entry is ")
//             println(entry)
//             btcPrice := satsToBitcoin(entry)
//             size := calculateOrderSizeFromPrecision(pairs[reversedIndex], calculateNumberOfCoinsToBuy(entry)/2)        
//             println(size)
//             limitOrder(client, direction, getTradingSymbol(pairs[reversedIndex]), offset, size, fmt.Sprintf("%.8f", btcPrice))//calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, pairs[reversedIndex]))
//         }
//         // limitOrder(client, direction, pairs[reversedIndex], offset, size, //calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, pairs[reversedIndex]))
//     }

// func calculateOrderSizeFromPrecision(asset string, size float64) string {

/*
 * function marketOrders
 * params: client
 ************************
//  */
func marketOrders(client *binance.Client, direction binance.SideType) {
    for i, asset := range pairs {
        highEntry := entries[i*2 + 1]
        numCoins := calculateNumberOfCoinsToBuy(highEntry)
        size := calculateOrderSizeFromPrecision(asset, numCoins*2) 
        if !ifOrderSizeMeetsMinimum(numCoins, highEntry) {
            continue
        }                       
        marketOrder(client, direction, getTradingSymbol(asset), size)
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
    // openPositions[asset] -= positionMultiplier * positionSizes[asset]        
    fmt.Println(order)
    return nil
}
