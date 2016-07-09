package smartcard

import (
    "fmt"
    "testing"
)

var CMD_SELECT = []byte{
    0x00, 0xA4, 0x04, 0x00, 0x08,
    0x90, 0x72, 0x5A, 0x9E, 0x3B, 0x10, 0x70, 0xAA,
}

var CMD_10 = []byte{
    0x00, 0x10, 0x00, 0x00, 0x0B,
}

func printHex(buffer []byte) {
    for _, b := range buffer {
        fmt.Printf("%02x", b)
    }
    fmt.Println("")
}

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
    printHex(CMD_SELECT)
    response, err := card.Transmit(CMD_SELECT)
    if err != nil { t.Error(err); return }
    printHex(response)

    fmt.Println("\nSend CMD 10")
    fmt.Println("-----------\n")
    printHex(CMD_10)
    response, err = card.Transmit(CMD_10)
    if err != nil { t.Error(err); return }
    printHex(response)
    fmt.Printf("Quoth the Applet, \"%s\"\n", string(response[:len(response)-2]))
}
