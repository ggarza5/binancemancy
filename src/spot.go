package main

import (
	"context"
	_ "encoding/json"
	"fmt"
	_ "github.com/ggarza5/binancemancy/execution"
	"github.com/ggarza5/go-binance-margin"
	"github.com/ggarza5/go-binance-margin/common"
	"github.com/pborman/getopt/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// "go.mongodb.org/mongo-driver/bson/primitive/objectid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//TODO: ReduceTrade Dust prevention
//TODO: correct calls that forget to add trailing 0s to the sat numbers for entry/exits

var (
	mongoUsername            = "admin"
	mongoHost1               = "cluster0-iy9i2.mongodb.net/test?authSource=admin&replicaSet=Cluster0-shard-0&readPreference=primary&ssl=true"
	mongoPw                  = "admin"
	mongoTradeCollectionName = "altBtcTrades"
	apiKey                   = "jIyd39L4YfD5CRvygwh5LY1IVilQ38NXY5RshUxKGwR1Sjj6ZGzynkxfK1p2jX0c"
	secretKey                = "3IbVAdTpwMN417BNbiwxc63NMpm0EZiBRbC7YFol4gbMytV4FxtfBfJ5dGkgq5Z2"
	openDirection            = "SELL"
	tradeDirection           = ""
	globalOffset             = 0.01
	positionMultiplier       = 1.0

	globalOffsetFlag       = "0.01"
	positionMultiplierFlag = "1.0"
	mode                   = "market"

	client     = binance.NewClient(apiKey, secretKey)
	priceSlice = []*binance.SymbolPrice{}
	prices     = make(map[string]string)

	pairs       = []string{}
	pairsFlag   = ""
	entriesFlag = ""
	entries     = []string{"0.0001"}

	closeAllFlag = ""
	closeAll     = false

	slFlag = ""
	sl     = "0"
	tpFlag = ""
	tps    = []string{}

	minimumOrderSize = 0.0001
	positionSize     = 0.006

	positionSizes = map[string]float64{
		"ICX":   1,
		"MATIC": 1,
	}
)

type Trade struct {
	Id         primitive.ObjectID  `json:"_id" bson:"_id"`
	Sl         string              `json:"sl" bson:"sl"`
	Tps        string              `json:"tp" bson:"tp"`
	Open       bool                `json:"open" bson:"open"`
	NumCoins   float64             `json:"numCoins" bson:"numCoins"`
	LastUpdate primitive.Timestamp `json:"lastUpdate" bson:"lastUpdate"`
	Multiplier float64             `json:"multiplier" bson:"multiplier"`
	Pair       string              `json:"pair" bson:"pair"`
	Entry      string              `json:"entry" bson:"entry"`
}

//Done:
//Entries

//Coin discovery process - get all binance balances, and if we find a coin that is not stored in mongo, discover the precision used for prices and trade amounts by trading with high precision to low precision until system allows trade

//2 options for SL
//Set limit orders to enter, then query those orders every minute or 5 minutes, then once they are filled, place stop orders
//enter with market orders and place the stop loss right away
//time-based stops

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

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
	getopt.FlagLong(&closeAllFlag, "close all", 'c', "close all").SetOptional()
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
	// println(tradeDirection)
	// println(globalOffsetFlag)
	// println(positionMultiplierFlag)
	// println(mode)
	// println(pairsFlag)
	// println(entriesFlag)
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
	if len(closeAllFlag) != 0 {
		if closeAllFlag[0] == ' ' || closeAllFlag[0] == '=' {
			closeAllFlag = closeAllFlag[1:]
		}
		if closeAllFlag[0] == 'f' || closeAllFlag[0] == 'F' {
			closeAll = false
		} else {
			closeAll = true
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
			sl = slFlag
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
	} else if tradeDirection[0] == ' ' || tradeDirection[0] == '=' {
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
	} else if mode[0] == 'r' || mode[0] == 'R' {
		mode = "reduce"
	} else if mode[0] == 'b' || mode[0] == 'B' {
		mode = "bot"
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
	} else {
		// if mode[0:2] == "ca" || mode[0:2] == "Ca" || mode[0:2] == "CA" || mode[0:2] == "cA" {
		if strings.ToLower(mode[0:2]) == "ca" {
			mode = "cancel"
			println("we set mode to cancel")
		} else if mode[0:2] == "cl" || mode[0:2] == "Cl" || mode[0:2] == "CL" || mode[0:2] == "cL" {
			mode = "close"
		} else {
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

func getPrefix(str string) string {
	strLen := len(str)
	return str[:strLen-4]
}

func trimTrailingZeros(str string) string {
	return strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(str, "000"), "00"), "0")
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

func getNegativeOrderOfMagnitude(price string) int {
	order := 0
	priceFloat := parseFloatHandleErr(price)
	for order = 0; priceFloat < 1.0; priceFloat *= 10 {
		order = order - 1
	}
	return order
}

//if passed a whole number = want order of magnitude
//if passed a decimal: want to make it a whole number then get order of magnitude
func getPositiveOrderOfMagnitude(price string) int {
	order := 0
	priceFloat := parseFloatHandleErr(price)
	for order = 0; priceFloat > 1.0; priceFloat /= 10 {
		order = order + 1
	}
	return order
}

func getNumSigFigs(price string) int {
	priceFloat := parseFloatHandleErr(price)
	if priceFloat > 0 {
		return getPositiveOrderOfMagnitude(price)
	} else {
		wholeNumberPrice := float64(10^(getNegativeOrderOfMagnitude(price)*-1)) * priceFloat
		wholeNumberPriceString := strconv.FormatFloat(wholeNumberPrice, 'E', -1, 64)
		return getPositiveOrderOfMagnitude(wholeNumberPriceString)
	}
}

// func fixForWholeNumber(price string, entry float) {
// 	mostSignificantOom := getNegativeOrderOfMagnitude(price)
// 	sigFigs := getNumSigFigs(entry)
// 	numLeadingZeros := mostSignificantOom*-1 - 1
// 	zerosToAdd := common.PricePrecisions[pairs[0]] - numLeadingZeros - sigFigs
// }

// func fixForDecimal(price string, entry float) {

// }

func addZeros(zerosToAdd int) {
	for i, e := range entries {
		fl, _ := strconv.ParseInt(e, 10, 64)
		entries[i] = strconv.FormatInt(fl*int64(math.Pow10(zerosToAdd)), 10)
	}
	for i, t := range tps {
		fl, _ := strconv.ParseInt(t, 10, 64)
		tps[i] = strconv.FormatInt(fl*int64(math.Pow10(zerosToAdd)), 10)
	}
	fl, _ := strconv.ParseInt(slFlag, 10, 64)
	sl = strconv.FormatInt(fl*int64(math.Pow10(zerosToAdd)), 10)
}

func getNumLeadingZeros(price string) int {
	mostSignificantOom := getNegativeOrderOfMagnitude(price)
	println(mostSignificantOom)
	numLeadingZeros := mostSignificantOom*-1 - 2
	return numLeadingZeros
}

func checkAndFixPrices(price string) {
	//Z = numReq - numIncluded
	//numReq = pricePrec - most significant oom - pricePrecision
	//price precision = price num leading zeros + numDigitsPassed + numToAdd
	priceFloat := parseFloatHandleErr(price)
	if priceFloat > 1 {
		return
	}
	// entryFloat := parseFloatHandleErr(entries[0])
	// if entryFloat > 1 {
	// 	fixForWholeNumber(price, entryFloat)
	// } else {
	// 	fixForDecimal(price, entryFloat)
	// }
	mostSignificantOom := getNegativeOrderOfMagnitude(price)
	sigFigs := getNumSigFigs(entries[0])
	numLeadingZeros := mostSignificantOom*-1 - 1
	zerosToAdd := common.PricePrecisions[pairs[0]] - numLeadingZeros - sigFigs
	addZeros(zerosToAdd)
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
	printFlagArguments()
	account := getAccount(client)

	if len(pairs) >= 1 {
		priceSlice, _ = client.NewListPricesService().Symbol(strings.Join(Map(pairs, getTradingSymbol), ",")).Do(context.Background())
	} else {
		priceSlice, _ = client.NewListPricesService().Do(context.Background())
	}
	priceSlice = Filter(priceSlice, isSymbolABTCTicker)
	fillPriceMapFromTickers()
	openSpotPositions := getOpenSpotPositions(account.Balances)

	if mode == "bot" {
		executeBotLogic(openSpotPositions)
	} else if mode == "reduce" {
		reduceBotPositions(openSpotPositions)
	} else if mode == "market" {
		marketOrders(client, stringToSide(tradeDirection))
	} else if mode == "limit" {
		//if we don't have 0.006 BTC, just return. That is currently 2% of account size
		if openSpotPositions["BTC"] < 0.004 {
			println("We did not have enough BTC to enter a new position.")
			return
		}
		// if len(entries) != 0 {
		// 	limitOrders()
		// }
		limitOrders(client, stringToSide(tradeDirection))
	} else if mode == "account" {
		account := getAccount(client)
		println(account)
	} else if mode == "cancel" {

		if len(pairsFlag) > 0 {
			openOrdersAcrossAllPairs := []*binance.Order{}

			for _, asset := range pairs {
				openOrders, err := client.NewListOpenOrdersService().Symbol(getTradingSymbol(asset)).Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
				for _, order := range openOrders {
					// fmt.Println(order)
					openOrdersAcrossAllPairs = append(openOrdersAcrossAllPairs, order)
				}
			}
			cancelSpotOrders(openOrdersAcrossAllPairs)
		} else {
			//TODO: Convert to floatCopy that just has all the trading pairs

			//cancel the open spot orders
			cancelSpotOrders(getOpenSpotOrders(openSpotPositions))
		}
	} else if mode == "close" {

	} else if mode == "server" {
		setupServer()
	} else {
		log.Fatal("Utility not called with a suitable argument to the mode flag. Exiting without execution.")
	}
}

func reduceBotPositions(openPositions map[string]float64) {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cur := getOpenTradeCursor()
	defer cur.Close(ctx)
	trade := Trade{}
	trades := []Trade{}
	_, _ = trade, trades
	for cur.Next(ctx) {
		// var result bson.M
		err := cur.Decode(&trade)
		if err != nil {
			println("error from cur.Decode")
			log.Fatal(err)
		}

		//determine whether a trade already hit stop loss or not
		stopped := determineWhetherStopped(openPositions, trade.Pair) //unmarshalString(trade.Pair))

		if stopped {
			//set position open to false
			setOpenPositionToFalse(trade.Pair)
			continue
		}

		//multiply numCoins by positionMultiplier
		// newNumCoins := trade.EumCoins * positionMultiplier
		newMultiplier := trade.Multiplier * positionMultiplier

		//get open orders on the pair; if none, position was continued upon in previous step
		//cancel orders
		openOrders := getOpenSpotOrdersForASinglePair(trade.Pair)
		// time.Sleep(2 * time.Second)
		cancelSpotOrders(openOrders)

		//Once orders are cancelled, position should be noted as closed in the database
		setOpenPositionToFalse(trade.Pair)

		//get original number of orders -- number of TPs
		//arithmetic involves "2" to deal with OCO orders
		originalNumOrders := len(strings.Split(trade.Tps, ",")) * 2
		numFilledOrders := (len(openOrders) - originalNumOrders) / 2
		//need to reduce current gross exposure -- not original
		//to do so -- get proportion of coins already sold -- from number of TPs hit

		//reduce position by multiplier
		remainingProportionOfGrossExposure := 1.0 - (float64(numFilledOrders) / float64(originalNumOrders))
		remainingCoins := float64(trade.NumCoins) * remainingProportionOfGrossExposure
		size := calculateOrderSizeFromPrecision(pairs[0], remainingCoins, positionMultiplier)
		sizeFloat := parseFloatHandleErr(size)
		println("we have this many coins")
		fmt.Println(remainingCoins)
		println("we are tryna sell this many coins")
		fmt.Println(size)
		marketOrder(client, stringToSide("SELL"), getTradingSymbol(pairs[0]), size)

		//Placement TODO: Dust prevention

		// coinPriceFloat := parseFloatHandleErr(prices[trade.Pair])
		// thresholdNumberOfDustCoins := minimumOrderSize / coinPriceFloat

		//round it down
		//calculate remainder of reducedTrade coins % 3 (if TPs are hit, then mod it by 2 or 1 since the closing orders
		//will be split into 1-3 TPs)

		//simply add this remainder to the market sell order which reduces numCoins in play, and subtract it from the
		// numCoins in mongo new trade. No need to mess with TPs because this is the remainder

		//Position
		//handle TPs
		entries = strings.Split(trade.Entry, ",")
		pairs = []string{trade.Pair}
		sl = trade.Sl
		trade.Tps = newTpsString(trade.Tps, numFilledOrders) //strings.Split(trade.Tps, ",")
		tps = strings.Split(trade.Tps, ",")
		for i, tp := range tps {
			if i < numFilledOrders {
				continue
			} else {
				handleTPWithoutOpening(tp, newMultiplier)
			}
		}
		addReducedTradeToMongo(trade, trade.NumCoins-sizeFloat, newMultiplier)
	}
	//after handling the new orders,
}

func newTpsString(passedTps string, numFilled int) string {
	switch numFilled {
	case 2:
		return strings.Join(strings.Split(passedTps, ",")[0:1], ",")
	case 1:
		return strings.Join(strings.Split(passedTps, ",")[0:2], ",")
	default:
		return passedTps
	}
}

func handleCloseLogic(openSpotPositions map[string]float64) {
	//cancel the open spot orders
	cancelSpotOrders(getOpenSpotOrders(openSpotPositions))

	//close the open spot positions entirely
	closeSpotPositions(openSpotPositions)
}

func handleTPWithoutOpening(tp string, passedMultiplier float64) {
	currentPrice := parseFloatHandleErr(prices[getTradingSymbol(pairs[0])])
	fmt.Println("jerere")
	//special case:
	//if 3 leading zeros and 4 sig figs, multiply by 10
	if satsToBitcoin(tp)  < currentPrice {
		fmt.Println(satsToBitcoin(tp))
		fmt.Println(currentPrice)
		log.Fatal("couldnt do this order")
	}
	highEntry := entries[len(entries)-1]
	numCoins := calculateNumberOfCoinsToBuy(highEntry)
	size := calculateOrderSizeFromPrecision(pairs[0], numCoins, passedMultiplier)
	btcPrice := satsToBitcoin(tp)
	o, err := client.NewCreateOCOService().Symbol(getTradingSymbol(pairs[0])).Side(binance.SideTypeSell).StopPrice(fmt.Sprintf("%.8f", satsToBitcoin(sl))).
		StopLimitPrice(fmt.Sprintf("%.8f", satsToBitcoin(sl))).Price(fmt.Sprintf("%.8f", btcPrice)).Quantity(size).StopLimitTimeInForce("GTC").Do(context.Background())
	handleError(err)
	fmt.Println(o)
}

//TODO: Need to integrate PricePrecisions into the sigFig and exponent calculation
func handleTP(tp string) {
	prec := common.PricePrecisions[pairs[0]]
	if prec == 0 {
		log.Fatal("If we don't get past this statement, you need to add the coin to PricePrecisions in go-margin/commons/helpers")
	}
	currentPrice := parseFloatHandleErr(prices[getTradingSymbol(pairs[0])])
	// println("line 618")
	// println(tp)
	// numLeadingZeros := getNumLeadingZeros(floatToString(satsToBitcoin(tp)))
	// println("num leading zeros", numLeadingZeros)
	// sigFigs := getNumSigFigs(floatToString(satsToBitcoin(tp)))
	// exponent := numLeadingZeros + sigFigs
	// //if numLeadingZeros + numSigFigs = 8, then forego the whole exponentiation stuff
	// tpF := parseFloatHandleErr(tp)
	if satsToBitcoin(tp)  < currentPrice {
		fmt.Println(satsToBitcoin(tp))
		fmt.Println(currentPrice)
		log.Fatal("couldnt do this order")
	}

	// if tpF / math.Pow10(exponent) < currentPrice {
	// 	// println(8 - numLeadingZeros - sigFigs)
	// 	// fmt.Println(float64((10^(8 - numLeadingZeros - sigFigs))))
	// 	fmt.Println(tpF / float64(math.Pow10(exponent)))
	// 	fmt.Println(currentPrice)
	// 	log.Fatal("couldnt do this order")
	// }

	highEntry := entries[len(entries)-1]
	// println(floatToString(parseFloatHandleErr(highEntry)/math.Pow10(8 - exponent)))
	// println(math.Pow10(8 - exponent))
	// numCoins := calculateNumberOfCoinsToBuy(floatToString(parseFloatHandleErr(highEntry)*math.Pow10(8 - exponent)))
	numCoins := calculateNumberOfCoinsToBuy(highEntry)
	println("numCoinsToBuy", numCoins)
	size := calculateOrderSizeFromPrecision(pairs[0], numCoins, positionMultiplier)
	// println("in handletp")
	marketOrder(client, stringToSide("BUY"), getTradingSymbol(pairs[0]), size)
	// println("in handletp")
	btcPrice := satsToBitcoin(tp)
	o, err := client.NewCreateOCOService().Symbol(getTradingSymbol(pairs[0])).Side(binance.SideTypeSell).StopPrice(fmt.Sprintf("%.8f", satsToBitcoin(sl))).
		StopLimitPrice(fmt.Sprintf("%.8f", satsToBitcoin(sl))).Price(fmt.Sprintf("%.8f", btcPrice)).Quantity(size).StopLimitTimeInForce("GTC").Do(context.Background())
	handleError(err)
	fmt.Println(o)
}

func handleTPs() {
	for _, tp := range tps {
		handleTP(tp)
	}
}

func getMongoClient() *mongo.Client {
	// mongoURI := fmt.Sprintf("mongodb+srv://%s:%s@%s", mongoUsername, mongoPw, mongoHost1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, getClientOptions())
	handleError(err)

	return client
}

func getTradeCollection() *mongo.Collection {
	return getMongoClient().Database("dolphin").Collection(mongoTradeCollectionName)
}

func getOrderSize(asset string, size float64, passedMultiplier float64) float64 {
	numCoins := calculateOrderSizeFromPrecision(asset, size, passedMultiplier)
	tbr := parseFloatHandleErr(numCoins)
	return tbr
}

func addTradeToMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	highEntry := entries[1]
	numCoins := calculateNumberOfCoinsToBuy(highEntry)
	// size := calculateOrderSizeFromPrecision(pairs[0], numCoins, positionMultiplier)
	size := getOrderSize(pairs[0], numCoins, positionMultiplier)
	if !ifOrderSizeMeetsMinimum(numCoins, highEntry) {
		log.Fatal("Trade order size was not high enough")
		return
	}
	res, err := getTradeCollection().InsertOne(ctx, bson.M{
		"pair":       pairs[0],
		"entry":      entriesFlag,
		"sl":         slFlag,
		"tp":         tpFlag,
		"open":       true,
		"numCoins":   size * float64(len(tps)),
		"multiplier": positionMultiplier,
		"lastUpdate": primitive.Timestamp{T: uint32(time.Now().Unix())},
	})
	handleError(err)
	if err.Error() != "" {
		println(err.Error())
		fmt.Println(res)
	}
}

func addReducedTradeToMongo(trade Trade, numCoinsArgument float64, multiplierArgument float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := getTradeCollection().InsertOne(ctx, bson.M{
		"pair":       trade.Pair,
		"entry":      trade.Entry,
		"sl":         trade.Sl,
		"tp":         trade.Tps,
		"open":       true,
		"numCoins":   numCoinsArgument,
		"multiplier": multiplierArgument,
		"lastUpdate": primitive.Timestamp{T: uint32(time.Now().Unix())},
	})
	handleError(err)
	if err.Error() != "" {
		println(err.Error())
		fmt.Println(res)
	}
}

func executeBotLogic(openSpotPositions map[string]float64) {
	//if we don't have 0.006 BTC, just return. That is currently 2% of account size
	if openSpotPositions["BTC"] < 0.004 {
		println("We did not have enough BTC to enter a new position.")
		return
	}
	if slFlag != "" {
		checkAndFixPrices(prices[getTradingSymbol(pairs[0])])

		// set OCO stop/TP orders
		handleTPs()

		//All TPs placed --> add to mongo
		addTradeToMongo()
	}
}

//TODO: Get list of coins NOT to close from file provided at execution time
func getOpenSpotPositions(balances []binance.Balance) map[string]float64 {
	openSpotPositions := make(map[string]float64)
	//get the pairs with open spot positions
	for _, balance := range balances {
		if isNonTradingCoin(balance.Asset) {
			continue
		}
		// if balance.Asset == "BTC" { continue }
		if balance.Asset == "USDT" {
			continue
		}
		if balance.Asset == "VTHO" {
			continue
		}

		if len(pairs) > 0 && !Includes(pairs, balance.Asset) && balance.Asset != "BTC" {
			continue
		}

		floatFree := parseFloatHandleErr(balance.Free)

		if balance.Asset == "BTC" {
			openSpotPositions[balance.Asset] = floatFree
			continue
		}

		floatLocked := parseFloatHandleErr(balance.Locked)
		if floatFree == 0 && floatLocked == 0 {
			continue
		}

		priceFloat := parseFloatHandleErr(prices[getTradingSymbol(balance.Asset)])
		if (floatFree+floatLocked)*priceFloat < minimumOrderSize {
			continue
		}
		openSpotPositions[balance.Asset] = floatFree + floatLocked

	}
	return openSpotPositions
}

func unmarshalString(data interface{}) string {
	var tbr string
	bson.Unmarshal(data.([]byte), &tbr)
	return tbr
}

func unmarshalFloat(data interface{}) float64 {
	// stringTbr := unmarshalString(data)
	var tbr float64
	bson.Unmarshal(data.([]byte), &tbr)
	return tbr
}

func closeSpotPositions(openPositions map[string]float64) {
	println("made it inside closeopen")
	if len(pairs) > 0 {
		for _, p := range pairs {
			println(p)
			tradingSymbol := getTradingSymbol(p)
			amt := calculateOrderSizeFromPrecision(tradingSymbol, openPositions[p], positionMultiplier)
			if amt == "0.0" {
				continue
			}
			spotCloseOrder, spotOrderErr := client.NewCreateOrderService().Symbol(tradingSymbol).
				Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
				Quantity(calculateOrderSizeFromPrecision(p, positionMultiplier*openPositions[p], positionMultiplier)).Do(context.Background())
			println(spotCloseOrder)
			if spotOrderErr != nil {
				fmt.Println(spotOrderErr)
			}
		}
	} else {
		openPositions = getOpenSpotPositions(getAccount(client).Balances)
		for k, v := range openPositions {
			if k == "BTC" {
				continue
			}
			println(getTradingSymbol(k))
			amt := calculateOrderSizeFromPrecision(k, v, positionMultiplier)
			if amt == "0.0" {
				continue
			}

			spotCloseOrder, spotOrderErr := client.NewCreateOrderService().Symbol(getTradingSymbol(k)).
				Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
				Quantity(calculateOrderSizeFromPrecision(k, v*positionMultiplier, positionMultiplier)).Do(context.Background())
			println(spotCloseOrder)
			if spotOrderErr != nil {
				fmt.Println(spotOrderErr)
			}
		}
	}
}

// func adjustPositionMultiplierForOrderState(numOrig int, numFilled int) {
// 	if numFilled == 0 || numOrig == 1 {
// 		return 1
// 	}
// 	if numOrig == 2 {
// 		if numFilled == 1  {
// 			return 0.5
// 		}
// 	} else if numOrig == 3 {
// 		if numFilled == 1 {

// 		} else if numFilled == 2 {

// 		}
// 	}
// }

func setOpenPositionToFalse(asset string) {
	var result Trade
	col := getTradeCollection()
	err := col.FindOne(context.TODO(), bson.M{"pair": asset, "open": true}).Decode(&result)
	handleError(err)
	filter := bson.M{"pair": asset, "open": true}
	result.Open = false
	// filter := bson.M{"pair": asset, "open": true}
	// update := bson.D{{"$set", bson.D{{"open", false}}}}
	_, err = col.ReplaceOne(context.Background(), filter, result)
	// _, err := getTradeCollection().UpdateOne(context.TODO(), filter, update, nil)
	handleError(err)
}

func determineWhetherStopped(openPositions map[string]float64, asset string) bool {
	if Includes(getMapKeys(openPositions), asset) {
		return false
	} else {
		return true
	}
}

func getClientOptions() *options.ClientOptions {
	mongoURI := fmt.Sprintf("mongodb+srv://%s:%s@%s", mongoUsername, mongoPw, mongoHost1)

	// Set client options and connect
	return options.Client().ApplyURI(mongoURI)
}

func getOpenTradeCursor() *mongo.Cursor {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cur, err := getTradeCollection().Find(ctx, bson.M{"open": true})
	handleError(err)
	return cur
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
	for _, order := range orders {
		_, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.TODO())
		if err != nil {
			fmt.Println(err)
			// return
		}
	}
}

func getOpenSpotOrdersForASinglePair(asset string) []*binance.Order {
	openOrders, err := client.NewListOpenOrdersService().Symbol(getTradingSymbol(asset)).Do(context.Background())
	handleError(err)
	return openOrders
}

func getOpenSpotOrders(openPositions map[string]float64) []*binance.Order {
	openOrdersAcrossAllPairs := []*binance.Order{}
	fmt.Println("adasd")
	if len(pairsFlag) > 0 {
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
		fmt.Println("adasd")
		for asset := range openPositions {
			println("getting open spot orders for " + asset)
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
	// println("in prob func")
	// println(u)
	floatU := parseFloatHandleErr(u)
	if floatU < 1 {
		return floatU
	}
	ii, _ := strconv.Atoi(u)
	// println(ii)
	return float64(ii) / math.Pow10(int(8))
}

func btcToSats(u string) int64 {
	ii, _ := strconv.ParseFloat(u, 64)
	return int64(float64(ii) * math.Pow10(int(8)))
}

func ifOrderSizeMeetsMinimum(numCoins float64, entry string) bool {
	entryNum, _ := strconv.ParseFloat(entry, 64)
	if numCoins*entryNum >= minimumOrderSize {
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
func limitOrders(client *binance.Client, direction binance.SideType) {
	for globalPairIndex := range pairs {
		reversedIndex := len(pairs) - globalPairIndex - 1
		for _, entry := range entries {
			btcPrice := satsToBitcoin(entry)
			numCoins := calculateNumberOfCoinsToBuy(entry) / 2
			if !ifOrderSizeMeetsMinimum(numCoins, entry) {
				continue
			}
			size := calculateOrderSizeFromPrecision(pairs[reversedIndex], calculateNumberOfCoinsToBuy(entry)/2, positionMultiplier)
			limitOrder(client, direction, getTradingSymbol(pairs[reversedIndex]), size, fmt.Sprintf("%.8f", btcPrice))
		}
	}
}

/*
 * function stopMarketOrders
 * params: client
 ************************

func stopMarketOrders(client *binance.Client, direction binance.SideType, offset float64, isReduceOnly binance.OrderReduceOnly, pairs ...string) {
    prices := getPrices(client)
    for globalPairIndex, _ := range pairs {
        reversedIndex := len(pairs) - globalPairIndex - 1
        size := calculateOrderSizeFromPrecision(pairs[reversedIndex], calculateNumberOfCoinsToBuy)
        stopMarketOrder(client, getOppositeDirection(direction), pairs[reversedIndex], size, calculateOrderPriceFromOffset(prices[reversedIndex].Price, offset, direction, pairs[reversedIndex]), isReduceOnly)
    }
}

/*
 * function stopMarketOrder
 * params: client
 ************************
 //error-returning
func stopMarketOrder(client *binance.Client, direction binance.SideType, asset string, size string, price string, isReduceOnly binance.OrderReduceOnly) error {
    order, err := client.NewCreateOrderService().Symbol(asset).
        Side(direction).Type(binance.OrderTypeStopLoss).
        Quantity(size).ReduceOnly(isReduceOnly).
        StopPrice(price).Do(context.Background())
    if err != nil {
        fmt.Println(err)
        return err
    }
    // openPositions[asset] -= positionMultiplier * positionSizes[asset]
    fmt.Println(order)
    return nil
}
*/

/*
 * function limitOrder
 * params: client, direction, asset, size, price
 ************************
 */
func limitOrder(client *binance.Client, direction binance.SideType, asset string, size string, price string) {
	println("calling limit order")
	println(price)
	println("asdasda")
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

func floatToString(floater float64) string {
	return fmt.Sprintf("%f", floater)
}

/*
 * function calculateOrderSizeFromPrecision
 * params: priceString string, offset float64, direction binance.SideType
 ************************
 */
func calculateOrderSizeFromPrecision(asset string, size float64, passedMultiplier float64) string {
	size = math.Floor(size*math.Pow10(common.PositionPrecisions[asset])) / float64(math.Pow10(common.PositionPrecisions[asset]))
	if common.PositionPrecisions[asset] == 0 {
		return fmt.Sprintf("%d", int64(passedMultiplier*size))
	} else if common.PositionPrecisions[asset] == 1 {
		return fmt.Sprintf("%.1f", passedMultiplier*size)
	} else if common.PositionPrecisions[asset] == 2 {
		return fmt.Sprintf("%.2f", passedMultiplier*size)
	} else if common.PositionPrecisions[asset] == 3 {
		return fmt.Sprintf("%.3f", passedMultiplier*size)
	} else if common.PositionPrecisions[asset] == 4 {
		return fmt.Sprintf("%.4f", passedMultiplier*size)
	} else if common.PositionPrecisions[asset] == 5 {
		return fmt.Sprintf("%.5f", passedMultiplier*size)
	} else if common.PositionPrecisions[asset] == 6 {
		return fmt.Sprintf("%.6f", passedMultiplier*size)
	} else if common.PositionPrecisions[asset] == 7 {
		return fmt.Sprintf("%.7f", passedMultiplier*size)
	} else {
		return fmt.Sprintf("%.8f", passedMultiplier*size)
	}
}

/*
 * function calculateOrderSizeFromPrecision
 * params: priceString string, offset float64, direction binance.SideType
 ************************
 */
func calculateOrderPriceFromPrecision(asset string, price float64) string {
	price = math.Floor(price*math.Pow10(common.PricePrecisions[asset])) / float64(math.Pow10(common.PricePrecisions[asset]))
	if common.PositionPrecisions[asset] == 0 {
		return fmt.Sprintf("%d", int64(price))
	} else if common.PositionPrecisions[asset] == 1 {
		return fmt.Sprintf("%.1f", price)
	} else if common.PositionPrecisions[asset] == 2 {
		return fmt.Sprintf("%.2f", price)
	} else if common.PositionPrecisions[asset] == 3 {
		return fmt.Sprintf("%.3f", price)
	} else if common.PositionPrecisions[asset] == 4 {
		return fmt.Sprintf("%.4f", price)
	} else if common.PositionPrecisions[asset] == 5 {
		return fmt.Sprintf("%.5f", price)
	} else if common.PositionPrecisions[asset] == 6 {
		return fmt.Sprintf("%.6f", price)
	} else if common.PositionPrecisions[asset] == 7 {
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

//get the number of coins to sell at each TP
func calculateNumberOfCoinsToBuy(price string) float64 {
	ff := parseFloatHandleErr(price)
	var numCoins float64
	if len(tps) == 2 {
		numCoins = ((3 * positionSize) / (ff * 2)) / 4
	} else if len(tps) == 3 {
		numCoins = (positionSize / ff) / 3
	} else { //there is only 1 TP left
		numCoins = positionSize / ff
	}
	return numCoins * math.Pow10(int(8))
}

/***
/*
 * function calculateOrderPriceFromOffset
 * params: priceString string, offset float64, direction binance.SideType
 ************************

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
//     if (common.PricePrecisions[asset] == 2) {
//         return fmt.Sprintf("%.2f", price - priceOffset)
//     } else if (common.PricePrecisions[asset] == 3) {
//         return fmt.Sprintf("%.3f", price - priceOffset)
//     } else if (common.PricePrecisions[asset] == 4) {
//         return fmt.Sprintf("%.4f", price - priceOffset)
//     } else {
//         return fmt.Sprintf("%.5f", price - priceOffset)
//     }
// }

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


 * function marketOrders
 * params: client
 ************************
//  */
func marketOrders(client *binance.Client, direction binance.SideType) {
	for i, asset := range pairs {
		highEntry := entries[i*2+1]
		numCoins := calculateNumberOfCoinsToBuy(highEntry)
		size := calculateOrderSizeFromPrecision(asset, numCoins, positionMultiplier)
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
 */ //error-returning
func marketOrder(client *binance.Client, direction binance.SideType, asset string, size string) *binance.Order {
	println(asset)
	order, err := client.NewCreateOrderService().Symbol(asset).
		Side(direction).Type(binance.OrderTypeMarket).
		Quantity(size).Do(context.Background())
	handleError(err)
	// openPositions[asset] -= positionMultiplier * positionSizes[asset]
	fmt.Println(order)
	return nil
}
