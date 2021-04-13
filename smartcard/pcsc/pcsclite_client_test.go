// +build !windows

package pcsc

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

func TestClient(t *testing.T) {
    fmt.Println("\n=================")
    fmt.Println("PCSCD Client Test")
    fmt.Printf("=================\n\n")
    fmt.Println("Connect to daemon")
    fmt.Printf("-----------------\n\n")
    client, err := PCSCLiteConnect()
    if err != nil { t.Error(err); return }
    defer client.Close()
    fmt.Println("OK")

    fmt.Println("\nEstablish Context")
    fmt.Printf("-----------------\n\n")
    context, err := client.EstablishContext()
    if err != nil { t.Error(err); return }
    defer client.ReleaseContext(context)
    fmt.Println("OK")

    fmt.Println("\nList Readers")
    fmt.Printf("------------\n\n")
    var selectedReader *Reader = nil
    readers, err := client.ListReaders()
    if err != nil { t.Error(err); return }
    for _, reader := range readers {
        fmt.Println(reader)
        if reader.IsCardPresent() && (selectedReader == nil) {
            selectedReader = reader
        }
    }

    if selectedReader == nil {
        fmt.Println("No reader with card found")
        return
    }

    fmt.Println("Connect to card")
    fmt.Printf("---------------\n\n")
    card, protocol, err := client.CardConnect( context, selectedReader.Name())
    if err != nil { t.Error(err); return }
    fmt.Println("OK")

    fmt.Println("\nSelect applet")
    fmt.Printf("-------------\n\n")
    buffer := make([]byte, 258)
    printHex(CMD_SELECT)
    received, err := client.Transmit(card, protocol, CMD_SELECT, buffer)
    if err != nil { t.Error(err); return }
    printHex(buffer[:received])

    fmt.Println("\nSend CMD 10")
    fmt.Printf("-----------\n\n")
    printHex(CMD_10)
    received, err = client.Transmit(card, protocol, CMD_10, buffer)
    if err != nil { t.Error(err); return }
    printHex(buffer[:received])
    fmt.Printf("Quoth the Applet, \"%s\"\n", string(buffer[:received-2]))

    fmt.Println("\nDisconnect from card")
    fmt.Printf("--------------------\n\n")
    err = client.CardDisconnect(card)
    if err != nil { t.Error(err); return }
    fmt.Println("OK")
}
