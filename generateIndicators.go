package main

import (
	"encoding/csv"
	"fmt"
	"github.com/ggarza5/technical-indicators"
	"github.com/pborman/getopt/v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

//TODO:
//Alert-server
//Trendline drawing service, wedge/chart pattern detection and classification
//Convexity detection
//Regresser
//Indicator-based bias classification
//Price-vol-time-based bias classifcation
//MLE calculator for forward-looking/now-casted time-series mean and standard deviation
//key level detectors - pivot points, horizontal levels that attracted price or required outsized volume or volatility to break

var (
	filename = ""
	out      = false
	symbol   = ""
)

/* function init
 * params:
 ************************
 * Initiates the global flag variables
 */
func init() {
	getopt.FlagLong(&filename, "fn", 'f', "filename").SetOptional()
	// getopt.FlagLong(&symbol, "sym", 's', "Symbol that we are generating indicators for").SetOptional()
	getopt.Flag(&out, 'o', "Whether to write results out to a file or not").SetOptional()
}

func main() {
	getopt.Parse()
	if filename == "" {
		filename = "BTCUSDT-1d-data.csv"
	}
	records, opens, highs, lows, closes := readInData(filename)
	conv, base, a, b, lag := generateIchi(highs, lows, closes)
	indicators := [][]float64{conv, base, a, b, lag}
	_, _ = records, opens

	// if symbol != "" {

	if out {
		bits := strings.Split(filename, "-")
		newFilename := bits[0] + "-" + bits[1] + "-indicators.txt"
		f, err := os.Create(newFilename)
		check(err)

		// w := io.NewWriter(f)
		fmt.Fprintln(f, indicators)
		defer f.Close()

		// w.WriteAll(indicators) // calls Flush internally

		// if err := w.Error(); err != nil {
		// 	log.Fatalln("error writing csv:", err)
		// }
	} else {
		for _, i := range indicators {
			fmt.Println(i)
		}
	}

	// }
}

//timestamp,open,high,low,close,volume,close_time,quote_av,trades,tb_base_av,tb_quote_av,ignore

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readInData(file string) ([][]string, []float64, []float64, []float64, []float64) {
	in, _ := ioutil.ReadFile(file)
	r := csv.NewReader(strings.NewReader(string(in)))
	var records [][]string
	var opens []float64
	var highs []float64
	var lows []float64
	var closes []float64
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		records = append(records, record)
		open, _ := strconv.ParseFloat(record[1], 64)
		high, _ := strconv.ParseFloat(record[2], 64)
		low, _ := strconv.ParseFloat(record[3], 64)
		close, _ := strconv.ParseFloat(record[4], 64)
		opens = append(opens, open)
		highs = append(highs, high)
		lows = append(lows, low)
		closes = append(closes, close)
	}
	return records, opens, highs, lows, closes
}

func generateIchi(highs []float64, lows []float64, closes []float64) ([]float64, []float64, []float64, []float64, []float64) {
	ichiParameters := []int{20, 60, 120, 30}
	conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan := indicators.IchimokuCloud(closes, lows, highs, ichiParameters)
	// _, _, _, _, _ = conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan
	return conversionLine, baseLine, leadSpanA, leadSpanB, lagSpan
	// fmt.Println(conversionLine[len(conversionLine)-20:])
	// fmt.Println(baseLine[len(baseLine)-20:])
	// fmt.Println(leadSpanA[len(leadSpanA)-20:])
	// fmt.Println(leadSpanB[len(leadSpanB)-20:])
	// fmt.Println(lagSpan[len(lagSpan)-20:])
}
