package main

import (
    "os"
    "fmt"
    "flag"
    "bufio"
    "encoding/hex"
    "strings"
    "github.com/sf1/go-card/smartcard"
)

func main() {
    var script, aid string
    flag.StringVar(&script, "script", "", "script")
    flag.StringVar(&aid, "aid", "", "applet id")
    flag.Usage = func() {
        fmt.Println("\nusage: apdu [ -aid <aid> -script <file> ]\n")
    }
    flag.Parse()
    err := run(aid, script)
    if err != nil {
        fmt.Printf("\nerror: %s\n\n", err)
        os.Exit(1)
    }
}

func run(aid, script string) error {
    var reader *smartcard.Reader
    ctx, err := smartcard.EstablishContext()
    if err != nil {
        return err
    }
    defer ctx.Release()
    readers, err := ctx.ListReadersWithCard()
    if err != nil {
        return err
    }
    if len(readers) == 0 {
        fmt.Println("\nplease insert smart card\n")
        return nil
    }
    if len(readers) == 1 {
        reader = readers[0]
    } else {
        // to do: handle multiple readers choices
        return fmt.Errorf("multiple readers not supported")
    }
    card, err := reader.Connect()
    if err != nil {
        return err
    }
    defer card.Disconnect()
    if aid != "" {
        fmt.Println("")
        err = processCommand(card, "select " + aid, true)
        if err != nil {
            return err
        }
    }
    if script == "" {
        return runInteractive(card)
    }
    return runScript(card, script)
}

func runInteractive(card *smartcard.Card) error {
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("\n>> ")
        scanner.Scan()
        command := scanner.Text()
        if command == "" {
            break
        }
        err := processCommand(card, command, false)
        if err != nil {
            fmt.Printf("error: %s\n", err)
            continue
        }
    }
    fmt.Println("")
    return nil
}

func processCommand(card *smartcard.Card, command string, echoCmd bool) error {
    apdu := make([]byte, 0, 128)
    parts := strings.Split(command, " ")
    isSelect := false
    if strings.ToLower(parts[0]) == "select" {
        isSelect = true
        parts = parts[1:]
        apdu = append(apdu, 0x00, 0xa4, 0x04, 0x00, 0x00)
    }
    for _, p := range parts {
        bytes, err := hex.DecodeString(p)
        if err != nil {
            return err
        }
        apdu = append(apdu, bytes...)
    }
    if isSelect {
        if len(apdu) == 5 {
            return fmt.Errorf("no aid provided")
        }
        apdu[4] = byte(len(apdu) - 5)
    }
    cmd := smartcard.CommandAPDU(apdu)
    if echoCmd {
        fmt.Printf(">> %s\n", cmd)
    }
    res, err := card.TransmitAPDU(cmd)
    if err != nil {
        return err
    }
    fmt.Printf("<< %s\n", res)
    return nil
}

func runScript(card *smartcard.Card, script string) error {
    return nil
}
