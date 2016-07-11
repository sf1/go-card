// +build windows

package smartcard

import (
    "fmt"
    "github.com/sf1/go-card/smartcard/pcsc"
)

type Context struct {
    ctxID uintptr
    winscard *pcsc.WinscardWrapper
}

func EstablishContext() (*Context, error) {
    winscard, err := pcsc.Winscard()
    if err != nil {
        return nil, err
    }
    ctxID, err := winscard.EstablishContext()
    if err != nil { return nil, err }
    return &Context{ctxID, winscard}, nil
}


func (ctx *Context) Release() error {
    return ctx.winscard.ReleaseContext(ctx.ctxID)
}

func (ctx *Context) ListReaders() ([]*Reader, error) {
    readerNames, err := ctx.winscard.ListReaders(ctx.ctxID)
    if err != nil { return nil, err }
    readers := make([]*Reader, len(readerNames))
    for i := 0; i < len(readerNames); i++ {
        readers[i] = &Reader{ctx, readerNames[i]}
    }
    return readers, nil
}

func (ctx *Context) ListReadersWithCard() ([]*Reader, error) {
    states, err := ctx.winscard.GetStatusChangeAll(
        ctx.ctxID, pcsc.SCARD_INFINITE, pcsc.SCARD_STATE_UNAWARE)
    if err != nil { return nil, err }
    readers := make([]*Reader, 0, len(states))
    for _, state := range states {
        if state.EventState & pcsc.SCARD_STATE_MUTE != 0 {
            continue
        }
        if state.EventState & pcsc.SCARD_STATE_PRESENT != 0 {
            readers = append(readers, &Reader{ctx, state.Reader})
        }
    }
    return readers, nil
}

func (ctx *Context) WaitForCardPresent() (*Reader, error) {
    var reader *Reader = nil
    states, err := ctx.winscard.GetStatusChangeAll(
        ctx.ctxID, pcsc.SCARD_INFINITE, pcsc.SCARD_STATE_UNAWARE)
    if err != nil { return nil, err }
    for reader == nil {
        for _, state := range states {
            state.CurrentState = state.EventState
            if state.EventState & pcsc.SCARD_STATE_MUTE != 0 {
                continue
            }
            if state.EventState & pcsc.SCARD_STATE_PRESENT != 0 {
                reader = &Reader{ctx, state.Reader}
                break
            }
        }
        if reader == nil {
            err = ctx.winscard.GetStatusChange(ctx.ctxID, pcsc.SCARD_INFINITE, states)
            if err != nil { return nil, err }
        }
    }
    return reader, nil
}

type Reader struct {
    context *Context
    name string
}

func (r *Reader) Name() string {
    return r.name
}

func (r *Reader) IsCardPresent() bool {
    states := make([]pcsc.ReaderState, 1)
    states[0].Reader = r.name
    states[0].CurrentState = pcsc.SCARD_STATE_UNAWARE
    err := r.context.winscard.GetStatusChange(r.context.ctxID,
        pcsc.SCARD_INFINITE, states)
    if err != nil {
        fmt.Println(err)
        return false
    }
    if states[0].EventState & pcsc.SCARD_STATE_MUTE != 0 {
        return false
    }
    return states[0].EventState & pcsc.SCARD_STATE_PRESENT != 0
}

func (r *Reader) Connect() (*Card, error) {
    var pci uintptr
    cardID, protocol, err := r.context.winscard.CardConnect(
        r.context.ctxID, r.name)
    if err != nil { return nil, err }
    switch(protocol) {
        case pcsc.SCARD_PROTOCOL_T0:
            pci = r.context.winscard.T0PCI()
        case pcsc.SCARD_PROTOCOL_T1:
            pci = r.context.winscard.T1PCI()
        default:
            return nil, fmt.Errorf("Unknown protocol: %08x", protocol)
    }
    return &Card{r.context, cardID, pci, nil}, nil
}

type Card struct {
    context *Context
    cardID uintptr
    sendPCI uintptr
    atr ATR
}

func (c *Card) Disconnect() error {
    err := c.context.winscard.CardDisconnect(c.cardID)
    if err != nil { return err }
    return nil
}

func (c *Card) ATR() ATR {
    var err error
    if c.atr != nil { return c.atr }
    c.atr, err = c.context.winscard.GetAttrib(c.cardID, pcsc.SCARD_ATTR_ATR_STRING)
    if err != nil { return nil }
    return c.atr
}

func (c *Card) Transmit(command []byte) ([]byte, error) {
    response := make([]byte, 258)
    received, err := c.context.winscard.Transmit(c.cardID, c.sendPCI,
        command, response)
    if err != nil { return nil, err }
    return response[:received], nil
}
