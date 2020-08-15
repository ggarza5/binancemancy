#!/usr/bin/env python3
import sys, os, argparse
import bitmex_and_binance
from bitmex_and_binance import get_all_binance
# sys.path.append('/Users/gabrielgarza/Developer/web/go/src/github.com/ggarza5/binancemancy')

# parser = argparse.ArgumentParser(description='Process some integers.')
# parser.add_argument('--file', dest='accumulate', action='store_const',
#                    const=sum, default=max,
#                    help='sum the integers (default: find the max)')

# args = parser.parse_args()

print(sys.argv[1])
# For Binance
binance_symbols = ["EOSBTC"]
if len(sys.argv) == 1:
	binance_symbols = ["EOSBTC"]
else:
	binance_symbols = [sys.argv[1]]
for symbol in binance_symbols:
    get_all_binance(symbol, '1d', True)
    get_all_binance(symbol, '1h', True)
