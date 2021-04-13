package main

import (
    "fmt"
    "github.com/sf1/go-card/smartcard"
)

func main() {
    ctx, err := smartcard.EstablishContext()
    if err != nil { panic(err) }
    defer ctx.Release()
    fmt.Printf("\nWaiting for card...")
    reader, err := ctx.WaitForCardPresent()
    if err != nil { panic(err) }
    card, err := reader.Connect()
    if err != nil { panic(err) }
    fmt.Printf("\n\nATR: %s\n\n",card.ATR())
    card.Disconnect()
    fmt.Printf("Please remove card");
    reader.WaitUntilCardRemoved()
    fmt.Printf("\n\n")
}
