package pcsclite

import (
    "fmt"
    "testing"
)

var cmdSelect = []byte{
    0x00, 0xA4, 0x04, 0x00, 0x08,
    0x90, 0x72, 0x5A, 0x9E, 0x3B, 0x10, 0x70, 0xAA,
}

var cmd10 = []byte{
    0x00, 0x10, 0x00, 0x00, 0x0B,
}

func printHex(buffer []byte) {
    for _, b := range buffer {
        fmt.Printf("%02x", b)
    }
    fmt.Println("")
}

func TestClient(t *testing.T) {
    fmt.Println("\n=====================")
    fmt.Println("Low Level Client Test")
    fmt.Println("=====================\n")
    fmt.Println("Connect to daemon")
    fmt.Println("-----------------\n")
    client, err := PCSCDConnect()
    if err != nil { t.Error(err); return }
    defer client.Close()
    fmt.Println("OK")

    fmt.Println("\nEstablish Context")
    fmt.Println("-----------------\n")
    context, err := client.EstablishContext()
    if err != nil { t.Error(err); return }
    defer client.ReleaseContext(context)
    fmt.Println("OK")

    fmt.Println("\nList Readers")
    fmt.Println("------------\n")
    var selectedReader *ReaderInfo = nil
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
    fmt.Println("---------------\n")
    card, protocol, err := client.CardConnect( context, selectedReader.Name())
    if err != nil { t.Error(err); return }
    fmt.Println("OK")

    fmt.Println("\nSelect applet")
    fmt.Println("-------------\n")
    buffer := make([]byte, 258)
    printHex(cmdSelect)
    received, err := client.Transmit(card, protocol, cmdSelect, buffer)
    if err != nil { t.Error(err); return }
    printHex(buffer[:received])

    fmt.Println("\nSend CMD 10")
    fmt.Println("-----------\n")
    printHex(cmd10)
    received, err = client.Transmit(card, protocol, cmd10, buffer)
    if err != nil { t.Error(err); return }
    printHex(buffer[:received])
    fmt.Printf("Quoth the Applet, \"%s\"\n", string(buffer[:received-2]))

    fmt.Println("\nDisconnect from card")
    fmt.Println("--------------------\n")
    err = client.CardDisconnect(card)
    if err != nil { t.Error(err); return }
    fmt.Println("OK")
}
