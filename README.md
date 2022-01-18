# binancemancy
This is Gabriel's repo for component files of the binance automated trading platform. There are master executables that include utilities for setting market orders, limit orders, pulling open positions, and soon setting stop losses. All the API keys are broken, so just replace them with yours if you want to use this library.
## Usage
This library is called from the bin folder using commands in the form of ./futures -ds -Mm -m1.0
This demonstrates the default use case of executing an index fund-like futures bet in the direction specifided by the d parameter, in a specified mode using the -M parameter (in this case, for mrket orrders), and ith a specifided multiplier (in this cas 1.0, using the -m paraeter).
There is another default use case shown in the file which is to run a server that listens to the price series evolution for pairs on Binance, and responds based on price action. This is the beginning of n automated trading utility, and the other functions should be understood as primitives for its functionality. More development coming soon.
### TODO
There are TODOs strewn through the codebase for improvements and in progresison toward further modularization of different automated trading use cases
What I need to do is split all of the utilities into different files so that they can be called from the server files and interface with each other upon different algorithmic triggers.