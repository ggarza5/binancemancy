package detection

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

//TODO
//calculate trendlines
//use intercept and slope, and then implement "go with market" strategy as structure gains uncertainty

func check(e error) {
	if e != nil {
		panic(e)
	}
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

//TODO: Kline version
//classify trendline as valid/regress validity
//compare trendlines

//Trend-scoring
//Quantitate the upward drift of securities
//Quantify the likelihood for full retraces

func detectDowntrend() {
	//band
	//resistance tapped
}

func detectUptrend() {
	//band
	//support tapped
	//

}

func detectBottom(timeframe string, data []int) {
	//bolliner band
	//detect support
	//ichimoku
	//Trendline
	//s/r level
}

func detectTop(timeframe string, data []int) {
	//bollinger band
	//detect resistance
	//ichimoku
	//Trendline
	//s/r level
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
//Detect volatility expansion and contraction from different reference points
//Calculate third derivative of price - ie the volatility of volatility
func calcualteVolatilityScore() {

}

//@version=4
// study(title="Historical Volatility", shorttitle="HV", format=format.price, precision=2, resolution="")
// length = input(10, minval=1)
// annual = 365
// per = timeframe.isintraday or timeframe.isdaily and timeframe.multiplier == 1 ? 1 : 7
// hv = 100 * stdev(log(close / close[1]), length) * sqrt(annual / per)
// plot(hv, "HV", color=#3A85AD)
