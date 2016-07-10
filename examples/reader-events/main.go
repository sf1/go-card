package main

import (
    "fmt"
    "github.com/sf1/go-card/smartcard"
)

func printReaderStates(ctx smartcard.Context) {
    readers, err := ctx.ListReaders()
    if err != nil { panic(err) }
    for _, reader := range readers {
        fmt.Printf("%s\n", reader.Name())
        fmt.Printf("- Card present: %t\n\n", reader.IsCardPresent())
    }
}

func main() {
    ctx, err := smartcard.EstablishContext()
    if err != nil { panic(err) }
    defer ctx.Release()
    fmt.Println("\nReader List")
    fmt.Println("-----------\n")
    printReaderStates(ctx)
    fmt.Println("Reader Events")
    fmt.Println("-------------\n")
    fmt.Println("Waiting...\n")
}
