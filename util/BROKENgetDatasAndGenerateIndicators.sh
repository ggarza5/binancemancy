#!/bin/bash
variable="$1"
for i in ${variable//,/ }
	python3 get_binance.py "$i"
	go run conv1hTo4h.go -f"$i"-1h-data.csv
	echo go run generateIndicators.go -f"$i"-4h-data.csv
	go run generateIndicators.go -f"$i"-4h-data.csv
done