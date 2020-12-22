package main

import (
	"fs"
	"http"
	"log"
)

func SetupServer() {
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.HandleFunc("/", serveTemplate)
    http.HandleFunc("/market_buy.json", marketBuyHandler)
    http.HandleFunc("/market_sell.json", marketSellHandler)
    http.HandleFunc("/limit_buy.json", limitBuyHandler)
    http.HandleFunc("/limit_sell.json", limitSellHandler)
    http.HandleFunc("/close.json", closeHandler)
    http.HandleFunc("/cancel.json", cancelHandler)
    log.Println("Listening...")
    http.ListenAndServe(":8080", nil)
}

/*
 * function HelloServer
 * params: client
 ************************
 */
func HelloServer(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!<br><button><a href='/'>Buy</a></button><br><button><a href='/'>Sell</a></button><br><button><a href='/'>Close</a></button><br>", r.URL.Path[1:])
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
    lp := filepath.Join("templates", "layout.html")
    fp := filepath.Join("templates", filepath.Clean(r.URL.Path))
    // println(r.URL.Path)
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

func marketBuyHandler(w http.ResponseWriter, r *http.Request) {
    TradeDirection = "BUY"    
    MarketOrders(client, stringToSide(TradeDirection))    
    OpenDirection = "BUY"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func marketSellHandler(w http.ResponseWriter, r *http.Request) {
    TradeDirection = "SELL"
    MarketOrders(client, stringToSide(TradeDirection))        
    OpenDirection = "SELL"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func limitBuyHandler(w http.ResponseWriter, r *http.Request) {
    TradeDirection = "BUY"    
    LimitOrders(client, stringToSide(TradeDirection), globalOffset)    
    OpenDirection = "BUY"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func limitSellHandler(w http.ResponseWriter, r *http.Request) {
    TradeDirection = "SELL"    
    LimitOrders(client, stringToSide(TradeDirection), globalOffset)    
    OpenDirection = "SELL"
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func cancelHandler(w http.ResponseWriter, r *http.Request) {
    cancelOrders(client, getOrders(client))    
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func closeHandler(w http.ResponseWriter, r *http.Request) {
    closeOpenPositions(client)
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}