package execution

import (
	"fmt"
	"github.com/ggarza5/binancemancy/common"
	"github.com/ggarza5/go-binance-margin"
)

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
			// marketOrders(client, )
		}
		//start calculating Bollinger Bands
		if len(dummyCloses) >= 20 {
			middle, upper, lower := indicators.BollingerBands(dummyCloses, 20, 2.0)
			_, _, _ = middle, upper, lower
			// fmt.Println(middle)
			// fmt.Println(upper)
			// fmt.Println(lower)

			var dummyHighs []float64
			var dummyLows []float64
			for _, f := range dummyCloses {
				dummyHighs = append(dummyHighs, f+rand.Float64())
				dummyLows = append(dummyLows, f-rand.Float64())
			}

			// fmt.Println(dummyHighs)
			conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan := indicators.IchimokuCloud(dummyCloses, dummyLows, dummyHighs, []int{20, 60, 120, 30})
			_, _, _, _, _ = conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan
			// fmt.Println(conversionLine)
			// fmt.Println(baseLine)
			// fmt.Println(leadSpanA)
			// fmt.Println(leadSpanB)
			// fmt.Println(lagSpan)
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

//if numUnits timeframe closes below level, exit the fucking trade
func timeBasedStop(client *binance.Client, pair string, timeframe string, numSeconds int, level float64) {

}

func calculateBasisDerivative(basis []float64) float64 {
	if len(basis) < 2 {
		return 0
	}
	return basis[len(basis)-1] - basis[len(basis)-2]
}

//TODO: Kline version
func tl(p1 float64, p2 float64, timeDiff int) {

}

//Trendline drawing
//When deciding to draw from the initial candle's open or extreme
//Minimize the sum of the deltas between the candle lows and the trendline
//When trendline converges to basis, consider it broken
//If 2 consecutive candles close below the trendline, consider it broken on that timeframe

func detectHighVolatilityTime() {

}

//Clculate ATR
//calculate normal distribution
//Price slope
//Second derivative of price
//Pearson's R of trends
//Detect
func calcualteVolatilityScore() {

}

//@version=4
// study(title="Historical Volatility", shorttitle="HV", format=format.price, precision=2, resolution="")
// length = input(10, minval=1)
// annual = 365
// per = timeframe.isintraday or timeframe.isdaily and timeframe.multiplier == 1 ? 1 : 7
// hv = 100 * stdev(log(close / close[1]), length) * sqrt(annual / per)
// plot(hv, "HV", color=#3A85AD)
