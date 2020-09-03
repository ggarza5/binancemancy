package main

import (
    "context"
    "fmt"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    _ "go.mongodb.org/mongo-driver/mongo/readpref"
    "log"

    "time"
)

var username = "admin"
var host1 = "cluster0-iy9i2.mongodb.net/test?authSource=admin&replicaSet=Cluster0-shard-0&readPreference=primary&ssl=true"
var pw = "admin"

func handleError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func main() {

    ctx := context.TODO()

    // pw, ok := os.LookupEnv("MONGO_PW")
    // if !ok {
    //     fmt.Println("error: unable to find MONGO_PW in the environment")
    //     os.Exit(1)
    // }

    mongoURI := fmt.Sprintf("mongodb+srv://%s:%s@%s", username, pw, host1)
    // mongoURI := "mongodb+srv://%s:%s@cluster0-iy9i2.mongodb.net/test?authSource=admin&replicaSet=Cluster0-shard-0&readPreference=primary&ssl=true"

    fmt.Println("connection string is:", mongoURI)

    // Set client options and connect
    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.Connect(ctx, clientOptions)
    handleError(err)

    err = client.Ping(ctx, nil)
    handleError(err)

    collection := client.Database("dolphin").Collection("altBtcTrades0")
    // var records bson.M

    ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)
    cur, err := collection.Find(ctx, bson.M{"open": true})
    handleError(err)
    // defer cur.Close(ctx)
    for cur.Next(ctx) {
        var result bson.M
        err := cur.Decode(&result)
        if err != nil {
            log.Fatal(err)
        }
        // do something with result....
        fmt.Println(result["entry"])

    }

    // handleError(err)
    // for _, r := range records {
    //     fmt.Println(r)
    // }

    fmt.Println("Connected to MongoDB!")
}
