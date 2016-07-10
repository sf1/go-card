package main

import (
    "fmt"
    "github.com/sf1/go-card/smartcard"
)

func main() {
    ctx, err := smartcard.EstablishContext()
    if err != nil { panic(err) }
    defer ctx.Release()
    fmt.Println("\nWaiting for card...")
    reader, err := ctx.WaitForCardPresent()
    if err != nil { panic(err) }
    card, err := reader.Connect()
    if err != nil { panic(err) }
    defer card.Disconnect()
    fmt.Printf("\nATR: %s\n\n",card.ATR())
}
