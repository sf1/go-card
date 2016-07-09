package pcsclite

import (
    "fmt"
    "testing"
)

func TestHighLevelAPI(t *testing.T) {
    fmt.Println("\n===================")
    fmt.Println("High Level API Test")
    fmt.Println("===================\n")

    fmt.Println("\nWait for card present")
    fmt.Println("---------------------\n")
    reader, err := WaitForCardPresent()
    if err != nil { t.Error(err); return }
    fmt.Println("OK")

    fmt.Println("\nConnect to card")
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
