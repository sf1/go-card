/*
Package smartcard a portable high-level API for communicating with smart cards.

Example:

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
*/
package smartcard

import (
    "fmt"
    "bytes"
)

type ATR []byte

// Return string form of ATR.
func (atr ATR) String() string {
    var buffer bytes.Buffer
    for _, b := range atr {
        buffer.WriteString(fmt.Sprintf("%02x", b))
    }
    return buffer.String()
}

// Transmit command APDU to the card and return response.
func (c *Card) TransmitAPDU(cmd *CommandAPDU) (*ResponseAPDU, error) {
    bytes, err := c.Transmit(cmd.Bytes())
    if err != nil { return nil, err }
    return Response(bytes), nil
}

// ISO7816-4 command APDU.
type CommandAPDU struct {
    Header [4]byte // APDU header (CLA, INS, P1, P2 fields)
    Data []byte // APDU data
    Le byte // APDU Le field
}

// Create command APDU with CLA, INS, P1, P2 as specified.
// No command data, no response required.
func Command(cla, ins, p1, p2 byte) *CommandAPDU {
    cmd := new(CommandAPDU)
    cmd.Header[0] = cla
    cmd.Header[1] = ins
    cmd.Header[2] = p1
    cmd.Header[3] = p2
    return cmd
}

// Create command APDU with CLA, INS, P1, P2 as specified.
// Response of length Le required.
func Command2(cla, ins, p1, p2, le byte) *CommandAPDU {
    cmd := Command(cla, ins, p1, p2)
    cmd.Le = le
    return cmd
}

// Create command APDU with CLA, INS, P1, P2 and data as specified.
// No response required.
func Command3(cla, ins, p1, p2 byte, data []byte) *CommandAPDU {
    cmd := Command(cla, ins, p1, p2)
    cmd.Data = data
    return cmd
}

// Create command APDU with CLA, INS, P1, P2 and data as specified.
// Response of length Le required.
func Command4(cla, ins, p1, p2 byte, data []byte, le byte) *CommandAPDU {
    cmd := Command(cla, ins, p1, p2)
    cmd.Data = data
    cmd.Le = le
    return cmd
}

// Create ISO7816-4 SELECT FILE APDU.
func SelectCommand(aid ...byte) *CommandAPDU {
    cmd := new(CommandAPDU)
    cmd.Header = [4]byte{0x00, 0xA4, 0x04, 0x00}
    cmd.Data = aid
    return cmd
}

// Return byte form of APDU.
func (cmd *CommandAPDU) Bytes() []byte {
    buffer := new(bytes.Buffer)
    buffer.Write(cmd.Header[:])
    if cmd.Data != nil {
        buffer.WriteByte(byte(len(cmd.Data)))
        buffer.Write(cmd.Data)
    }
    if cmd.Le != 0 {
        buffer.WriteByte(cmd.Le)
    }
    return buffer.Bytes()
}

// Return string form of APDU.
func (cmd *CommandAPDU) String() string {
    buffer := new(bytes.Buffer)
    buffer.WriteString(
        fmt.Sprintf(
            "CLA: %02x, INS: %02x, P1: %02x, P2: %02x",
            cmd.Header[0],
            cmd.Header[1],
            cmd.Header[2],
            cmd.Header[3],
    ))
    if len(cmd.Data) > 0 {
        buffer.WriteString(fmt.Sprintf(", Lc: %02x, Data: ", len(cmd.Data)))
        for _, b := range cmd.Data {
            buffer.WriteString(fmt.Sprintf("%02x", b))
        }
    }
    if cmd.Le != 0 {
        buffer.WriteString(fmt.Sprintf("%02x", cmd.Le))
    }
    return buffer.String()
}

// ISO7816-4 response APDU.
type ResponseAPDU struct {
    Data []byte
    SW1 byte
    SW2 byte
}

// Create response APDU instance.
func Response(bytes []byte) *ResponseAPDU {
    n := len(bytes)
    if n < 2 {
        return nil
    }
    response := new(ResponseAPDU)
    response.SW2 = bytes[n-1]
    response.SW1 = bytes[n-2]
    if n > 2 {
        response.Data = bytes[:n-2]
    }
    return response
}

// Return 16-bit status word.
func (r *ResponseAPDU) SW() uint16 {
    return uint16(r.SW1) << 8 | uint16(r.SW2)
}

// Return string form of APDU.
func (r *ResponseAPDU) String() string {
    buffer := new(bytes.Buffer)
    if len(r.Data) > 0 {
        buffer.WriteString("Data: ")
        for _, b := range r.Data {
            buffer.WriteString(fmt.Sprintf("%02x", b))
        }
        buffer.WriteString(", ")
    }
    buffer.WriteString(fmt.Sprintf("SW: %04x", r.SW()))
    return buffer.String()
}
