package execution

import (
	"encoding/csv"
	"fmt"
	"github.com/ggarza5/go-binance-margin"
	_ "github.com/ggarza5/go-binance-margin/common"
	"github.com/ggarza5/technical-indicators"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func convert1hTo4hKlines(file string) {
	in, _ := ioutil.ReadFile(file)
	r := csv.NewReader(strings.NewReader(string(in)))
	var records [][]string
	var open string
	var high string
	var low string
	var close string
	//every 4 hours, update the timestamp for 4h kline
	counter := 0
	var date string
	r.Read()

	for {

		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if counter%4 == 0 {
			date = record[0]
			open = record[1]
			high = record[2]
			low = record[3]
			close = record[4]
		} else {
			if record[2] > high {
				high = record[2]
			}
			if record[3] < low {
				low = record[3]
			}
		}

		if counter%4 == 3 {
			//set closing price,
			//append kline to list,
			//continue
			close = record[4]
			kline := []string{date, open, high, low, close}
			records = append(records, kline)
		}
		counter = counter + 1
	}
	bits := strings.Split(file, "-")
	newFilename := bits[0] + "-4h-data.csv"
	// f, err := os.Create("BTCUSDT-4h-data.csv")
	f, err := os.Create(newFilename)
	check(err)

	defer f.Close()

	w := csv.NewWriter(f)
	w.WriteAll(records) // calls Flush internally

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
}

var dataCloses = make([]string, 0)
var dataClosesInts = make([]int, 0)
var dataClosesFloats = make([]float64, 32)
var klines = make([]*binance.WsKlineEvent, 0)

func listenOnSocket(client *binance.Client, pair string, timeframe string) {
	var opens []float64
	var highs []float64
	var lows []float64
	var closes []float64
	_ = opens
	_ = dataClosesInts
	_ = dataClosesFloats
	wsKlineHandler := func(event *binance.WsKlineEvent) {
		if event.Kline.IsFinal {
			// fmt.Println(event)
			klines = append(klines, event)
			dataCloses := append(dataCloses, event.Kline.Close)
			i, _ := strconv.Atoi(event.Kline.Close)
			dataClosesInts = append(dataClosesInts, i)
			dataClosesFloats = append(dataClosesFloats, float64(i))
			fmt.Println(dataCloses)
			// marketOrders(client, )
		}
		//start calculating Bollinger Bands
		if len(closes) >= 20 {
			middle, upper, lower := indicators.BollingerBands(closes, 20, 2.0)
			_, _, _ = middle, upper, lower
			// fmt.Println(middle)
			// fmt.Println(upper)
			// fmt.Println(lower)

			for _, f := range closes {
				highs = append(highs, f+rand.Float64())
				lows = append(lows, f-rand.Float64())
			}

			// fmt.Println(highs)
			conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan := indicators.IchimokuCloud(closes, lows, highs, []int{20, 60, 120, 30})
			_, _, _, _, _ = conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan
			// fmt.Println(conversionLine)
			// fmt.Println(baseLine)
			// fmt.Println(leadSpanA)
			// fmt.Println(leadSpanB)
			// fmt.Println(lagSpan)
		}
		// fmt.Println(klinesi)
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

//if numUnits timeframe closes below level, exit the fucking trade
/*
 * function TimeBasedStop
 * params: 
 ************************
 * Prints the global flag variables. Should be called after they are set by init()
 */
func TimeBasedStop(client *binance.Client, pair string, timeframe string, numSeconds int, level float64) {
	//start a timer when the price goes past a certain level (the opposite of the direction that the position is open in)
	//After the timer passes numSeconds, close the position or reduce exposure
}

/*
 * function ReduceExposure
 * params: 
 ************************
 * Prints the global flag variables. Should be called after they are set by init()
 */
func ReduceExposure(client *binance.Client, pair string, timeframe string, numSeconds int, level float64) {

	//Run through the open positions for client
	//Reduce all orders by percentage
	//market reduce all positions by percentage
}

/*
 * function IncreaseExposure
 * params: 
 ************************
 * Prints the global flag variables. Should be called after they are set by init()
 */
func IncreaseExposure(client *binance.Client, pair string, timeframe string, numSeconds int, level float64) {
	//Run through the open positions for client
	//Reduce all orders by percentage
	//market reduce all positions by percentage
}

/*
 * function WatchForShortTermDeviation
 * params: 
 ************************
 * Prints the global flag variables. Should be called after they are set by init()
 * 
 */
func WatchForShortTermDeviation(client *binance.Client, pair string, timeframe string, numSeconds int, level float64) {
	//get bollinger band for the timeframe
	//when the 
}












