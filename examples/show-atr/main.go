package main

import (
    "fmt"
    "github.com/sf1/go-sc/smartcard"
)

func main() {
    fmt.Println("\nWaiting for card...")
    reader, err := smartcard.WaitForCardPresent()
    if err != nil {
        panic(err)
    }
    card, err := reader.Connect()
    if err != nil {
        panic(err)
    }
    defer card.Disconnect()
    fmt.Printf("\nATR: %s\n\n",card.ATR())
}
