package smartcard

import (
    "fmt"
    "testing"
    "github.com/sf1/go-card/smartcard/pcsc"
)

func TestInfo(t *testing.T) {
    fmt.Println("\n===================")
    fmt.Println("High Level API Test")
    fmt.Printf("===================\n\n")
}

func TestEstablishReleaseUserContext(t *testing.T) {
    fmt.Println("------------------------------")
    fmt.Println("Test establish/release User Context")
    fmt.Printf("------------------------------\n\n")
    ctx, err := EstablishContext(pcsc.CARD_SCOPE_USER)
    if err != nil { t.Error(err); return }
    err = ctx.Release()
    if err != nil { t.Error(err); return }
    fmt.Printf("OK\n\n")
}

func TestEstablishReleaseSystemContext(t *testing.T) {
    fmt.Println("------------------------------")
    fmt.Println("Test establish/release System Context")
    fmt.Printf("------------------------------\n\n")
    ctx, err := EstablishContext(pcsc.CARD_SCOPE_SYSTEM)
    if err != nil { t.Error(err); return }
    err = ctx.Release()
    if err != nil { t.Error(err); return }
    fmt.Printf("OK\n\n")
}

func TestListReaders(t *testing.T) {
    fmt.Println("-----------------")
    fmt.Println("Test list readers")
    fmt.Printf("-----------------\n\n")
    ctx, err := EstablishContext(pcsc.CARD_SCOPE_SYSTEM)
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
    fmt.Printf("---------------------------\n\n")
    ctx, err := EstablishContext(pcsc.CARD_SCOPE_SYSTEM)
    if err != nil { t.Error(err); return }
    defer ctx.Release()
    readers, err := ctx.ListReadersWithCard()
    if err != nil { t.Error(err); return }
    for _, reader := range readers {
        fmt.Println(reader.Name())
        fmt.Printf("- Card present: %t\n\n", reader.IsCardPresent())
    }
}

func TestWaitForCardPresentRemoved(t *testing.T) {
    fmt.Println("----------------------------------------------------")
    fmt.Println("Test wait for card present / wait until card removed")
    fmt.Printf("----------------------------------------------------\n\n")
    ctx, err := EstablishContext(pcsc.CARD_SCOPE_SYSTEM)
    if err != nil { t.Error(err); return }
    defer ctx.Release()
    fmt.Printf("Insert card now...")
    reader, err := ctx.WaitForCardPresent()
    if err != nil { t.Error(err); return }
    fmt.Printf("\n\n%s\n", reader.Name())
    fmt.Printf("- Card present: %t\n\n", reader.IsCardPresent())
    fmt.Printf("Remove card now...")
    reader.WaitUntilCardRemoved()
    fmt.Printf("\n\nCard was removed\n\n")
}

func TestCardCommunication(t *testing.T) {
    fmt.Println("-----------------------")
    fmt.Println("Test card communication")
    fmt.Printf("-----------------------\n\n")
    ctx, err := EstablishContext(pcsc.CARD_SCOPE_SYSTEM)
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
    fmt.Printf("---------------\n\n")
    card, err := reader.Connect()
    if err != nil { t.Error(err); return }
    fmt.Printf("ATR: %s\n\n", card.ATR())

    fmt.Println("Select applet")
    fmt.Printf("-------------\n\n")
    cmd := SelectCommand(0x90, 0x72, 0x5A, 0x9E, 0x3B, 0x10, 0x70, 0xAA)
    fmt.Printf(">> %s\n", cmd)
    response, err := card.TransmitAPDU(cmd)
    if err != nil { t.Error(err); return }
    fmt.Printf("<< %s\n", response.String())

    fmt.Println("\nSend CMD 10")
    fmt.Printf("-----------\n\n")
    cmd = Command2(0x00, 0x10, 0x00, 0x00, 0x0b)
    fmt.Printf(">> %s\n", cmd)
    response, err = card.TransmitAPDU(cmd)
    if err != nil { t.Error(err); return }
    fmt.Printf("<< %s\n", response)
    fmt.Printf("\nQuoth the Applet, \"%s\"\n\n", string(response.Data()))

    fmt.Println("Disconnect from card")
    fmt.Printf("--------------------\n\n")
    err = card.Disconnect()
    if err != nil { t.Error(err); return }
    fmt.Printf("OK\n\n")
}
