package main

import (
	"encoding/csv"
	"fmt"
	"github.com/pborman/getopt/v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	_ "strconv"
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
)

/* function init
 * params:
 ************************
 * Initiates the global flag variables
 */
func init() {
	getopt.FlagLong(&filename, "fn", 'f', "Filename to be converted from 1h klines to 4h klines").SetOptional()
}

func main() {
	getopt.Parse()
	if filename == "" {
		filename = "BTCUSDT-1h-data.csv"
	}
	fmt.Println("Converting data in " + filename + " to 4h klines")
	convert1hTo4hKlines(filename)
	// convert1hTo4hKlines()
}

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
