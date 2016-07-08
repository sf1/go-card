package pcsclite

import (
    "fmt"
    "testing"
)

func TestHighLevelAPI(t *testing.T) {
    fmt.Println("\n===================")
    fmt.Println("High Level API Test")
    fmt.Println("===================\n")

    fmt.Println("Establish Context")
    fmt.Println("-----------------\n")
    ctx, err := EstablishContext()
    if err != nil { t.Error(err); return }
    defer ctx.Release()
    fmt.Println("OK")

    fmt.Println("\nList Readers")
    fmt.Println("------------\n")
    readers, err := ctx.ListReaders()
    if err != nil { t.Error(err); return }
    var reader *Reader = nil
    for _, r := range readers {
        fmt.Println(r.Name())
        fmt.Printf("- Card Present: %t\n\n", r.IsCardPresent())
        if r.IsCardPresent() && reader == nil {
            reader = r
        }
    }

    if reader == nil {
        fmt.Println("No reader with card found")
        return
    }

    fmt.Println("Connect to card")
    fmt.Println("---------------\n")
    card, err := reader.Connect()
    if err != nil { t.Error(err); return }
    defer card.Disconnect()
    fmt.Print("ATR: ")
    printHex(card.ATR())

    fmt.Println("\nSelect applet")
    fmt.Println("-------------\n")
    printHex(cmdSelect)
    response, err := card.Transmit(cmdSelect)
    if err != nil { t.Error(err); return }
    printHex(response)

    fmt.Println("\nSend CMD 10")
    fmt.Println("-----------\n")
    printHex(cmd10)
    response, err = card.Transmit(cmd10)
    if err != nil { t.Error(err); return }
    printHex(response)
    fmt.Printf("Quoth the Applet, \"%s\"\n", string(response[:len(response)-2]))
}
