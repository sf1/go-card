package smartcard

import (
    "fmt"
    "testing"
)

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
    fmt.Printf("ATR: %s\n\n", card.ATR())

    fmt.Println("Select applet")
    fmt.Println("-------------\n")
    cmd := SelectCommand(0x90, 0x72, 0x5A, 0x9E, 0x3B, 0x10, 0x70, 0xAA)
    fmt.Printf(">> %s\n", cmd)
    response, err := card.TransmitAPDU(cmd)
    if err != nil { t.Error(err); return }
    fmt.Printf("<< %s\n", response)

    fmt.Println("\nSend CMD 10")
    fmt.Println("-----------\n")
    cmd = Command2(0x00, 0x10, 0x00, 0x00, 0x0b)
    fmt.Printf(">> %s\n", cmd)
    response, err = card.TransmitAPDU(cmd)
    if err != nil { t.Error(err); return }
    fmt.Printf("<< %s\n", response)
    fmt.Printf("\nQuoth the Applet, \"%s\"\n\n", string(response.Data))

    fmt.Println("Disconnect from card")
    fmt.Println("--------------------\n")
    err = card.Disconnect()
    if err != nil { t.Error(err); return }
    fmt.Println("OK\n")
}
