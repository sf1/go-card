Go Smart Card API
=================

High level API for smart card integration written in pure Go. On Linx and other Unix-like systems, this package talks directly to the PCSC-Lite daemon, and on Windows it accesses the smart card service via winscard.dll.

Prerequisites
-------------

### Linux

Install the PCSC-Lite daemon and CCID driver.

Ubuntu:

    sudo apt-get install pcscd libccid

Arch Linux:

    sudo pacman -S pcsclite ccid
    sudo systemctl enable pcscd

### OS X

Probably none. PCSC-Lite should come with the system.

### Windows

None

Installation
------------

    go get github.com/sf1/go-card/smartcard

Usage
-----

    ctx, err := smartcard.EstablishContext()
    // handle error, if any
    defer ctx.Release()

    reader, err := ctx.WaitForCardPresent()
    // handle error, if any

    card, err := reader.Connect()
    // handle error, if any
    defer card.Disconnect()

    fmt.Printf("Card ATR: %s\n", card.ATR()) 
    command := SelectCommand(0xa0, 0x00, 0x00, 0x00, 0x62, 0x03, 0x01, 0xc, 0x01, 0x01)
    response, err := card.TransmitAPDU(command)
    // handle error, if any
    fmt.Printf("Response: %s\n", response)

License
-------

MIT. See [LICENSE](LICENSE)
