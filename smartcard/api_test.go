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

func TestInfo(t *testing.T) {
    fmt.Println("\n===================")
    fmt.Println("High Level API Test")
    fmt.Println("===================\n")
}

func TestEstablishReleaseContext(t *testing.T) {
    fmt.Println("------------------------------")
    fmt.Println("Test establish/release Context")
    fmt.Println("------------------------------\n")
    ctx, err := EstablishContext()
    if err != nil { t.Error(err); return }
    err = ctx.Release()
    if err != nil { t.Error(err); return }
    fmt.Println("OK\n")
}

func TestListReaders(t *testing.T) {
    fmt.Println("-----------------")
    fmt.Println("Test list readers")
    fmt.Println("-----------------\n")
    ctx, err := EstablishContext()
    if err != nil { t.Error(err); return }
    defer ctx.Release()
    readers, err := ctx.ListReaders()
    if err != nil { t.Error(err); return }
    for _, reader := range readers {
        fmt.Println(reader.Name())
        fmt.Printf("- Card present: %t\n\n", reader.IsCardPresent())
    }
}

func TestListReadersWithCard(t *testing.T) {
    fmt.Println("---------------------------")
    fmt.Println("Test list readers with card")
    fmt.Println("---------------------------\n")
    ctx, err := EstablishContext()
    if err != nil { t.Error(err); return }
    defer ctx.Release()
    readers, err := ctx.ListReadersWithCard()
    if err != nil { t.Error(err); return }
    for _, reader := range readers {
        fmt.Println(reader.Name())
        fmt.Printf("- Card present: %t\n\n", reader.IsCardPresent())
    }
}

func TestWaitForCardPresent(t *testing.T) {
    fmt.Println("--------------------------")
    fmt.Println("Test wait for card present")
    fmt.Println("--------------------------\n")
    ctx, err := EstablishContext()
    if err != nil { t.Error(err); return }
    defer ctx.Release()
    reader, err := ctx.WaitForCardPresent()
    if err != nil { t.Error(err); return }
    fmt.Println(reader.Name())
    fmt.Printf("- Card present: %t\n\n", reader.IsCardPresent())
}

func TestCardCommunication(t *testing.T) {
    fmt.Println("-----------------------")
    fmt.Println("Test card communication")
    fmt.Println("-----------------------\n")
    ctx, err := EstablishContext()
    if err != nil { t.Error(err); return }
    defer ctx.Release()
    readers, err := ctx.ListReadersWithCard()
    if err != nil { t.Error(err); return }
    if len(readers) == 0 {
        t.Error("No reader with card")
        return
    }
    reader := readers[0]
    fmt.Println("Connect to card")
    fmt.Println("---------------\n")
    card, err := reader.Connect()
    if err != nil { t.Error(err); return }
    fmt.Println("OK\n")

    fmt.Println("Select applet")
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
    fmt.Printf("Quoth the Applet, \"%s\"\n\n", string(response[:len(response)-2]))

    fmt.Println("Disconnect from card")
    fmt.Println("--------------------\n")
    err = card.Disconnect()
    if err != nil { t.Error(err); return }
    fmt.Println("OK\n")
}
