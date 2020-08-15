#!/bin/bash
python3 get_binance.py "$1"
go run conv1hTo4h.go -f"$1"-1h-data.csv
echo go run generateIndicators.go -f"$1"-4h-data.csv -o
go run generateIndicators.go -f"$1"-4h-data.csv -o