// +build windows

package smartcard

import (
    "fmt"
)

type WinscardContext struct {
    ctxID uintptr
    winscard *WinscardWrapper
}

func EstablishContext() (Context, error) {
    winscard, err := Winscard()
    if err != nil {
        return nil, err
    }
    ctxID, err := winscard.EstablishContext()
    if err != nil { return nil, err }
    return &WinscardContext{ctxID, winscard}, nil
}


func (ctx *WinscardContext) Release() error {
    return ctx.winscard.ReleaseContext(ctx.ctxID)
}

func (ctx *WinscardContext) ListReaders() ([]Reader, error) {
    readerNames, err := ctx.winscard.ListReaders(ctx.ctxID)
    if err != nil { return nil, err }
    readers := make([]Reader, len(readerNames))
    for i := 0; i < len(readerNames); i++ {
        readers[i] = &WinscardReader{ctx, readerNames[i]}
    }
    return readers, nil
}

func (ctx *WinscardContext) ListReadersWithCard() ([]Reader, error) {
    states, err := ctx.winscard.GetStatusChangeAll(
        ctx.ctxID, _SCARD_INFINITE, _SCARD_STATE_UNAWARE)
    if err != nil { return nil, err }
    readers := make([]Reader, 0, len(states))
    for _, state := range states {
        if state.EventState & _SCARD_STATE_MUTE != 0 {
            continue
        }
        if state.EventState & _SCARD_STATE_PRESENT != 0 {
            readers = append(readers, &WinscardReader{ctx, state.Reader})
        }
    }
    return readers, nil
}

func (ctx *WinscardContext) WaitForCardPresent() (Reader, error) {
    var reader *WinscardReader = nil
    states, err := ctx.winscard.GetStatusChangeAll(
        ctx.ctxID, _SCARD_INFINITE, _SCARD_STATE_UNAWARE)
    if err != nil { return nil, err }
    for reader == nil {
        for _, state := range states {
            state.CurrentState = state.EventState
            if state.EventState & _SCARD_STATE_MUTE != 0 {
                continue
            }
            if state.EventState & _SCARD_STATE_PRESENT != 0 {
                reader = &WinscardReader{ctx, state.Reader}
                break
            }
        }
        if reader == nil {
            err = ctx.winscard.GetStatusChange(ctx.ctxID, _SCARD_INFINITE, states)
            if err != nil { return nil, err }
        }
    }
    return reader, nil
}

type WinscardReader struct {
    context *WinscardContext
    name string
}

func (r *WinscardReader) Name() string {
    return r.name
}

func (r *WinscardReader) IsCardPresent() bool {
    states := make([]ReaderState, 1)
    states[0].Reader = r.name
    states[0].CurrentState = _SCARD_STATE_UNAWARE
    err := r.context.winscard.GetStatusChange(r.context.ctxID,
        _SCARD_INFINITE, states)
    if err != nil {
        fmt.Println(err)
        return false
    }
    if states[0].EventState & _SCARD_STATE_MUTE != 0 {
        return false
    }
    return states[0].EventState & _SCARD_STATE_PRESENT != 0
}

func (r *WinscardReader) Connect() (Card, error) {
    var pci uintptr
    cardID, protocol, err := r.context.winscard.CardConnect(r.context.ctxID, r.name)
    if err != nil { return nil, err }
    switch(protocol) {
        case _SCARD_PROTOCOL_T0:
            pci = r.context.winscard.T0PCI()
        case _SCARD_PROTOCOL_T1:
            pci = r.context.winscard.T1PCI()
        default:
            return nil, fmt.Errorf("Unknown protocol: %08x", protocol)
    }
    return &WinscardCard{r.context, cardID, pci}, nil
}

type WinscardCard struct {
    context *WinscardContext
    cardID uintptr
    sendPCI uintptr
}

func (c *WinscardCard) Disconnect() error {
    err := c.context.winscard.CardDisconnect(c.cardID)
    if err != nil { return err }
    return nil
}

func (c *WinscardCard) ATR() ATR {
    return nil
}

func (c *WinscardCard) Transmit(command []byte) ([]byte, error) {
    response := make([]byte, 258)
    received, err := c.context.winscard.Transmit(c.cardID, c.sendPCI,
        command, response)
    if err != nil { return nil, err }
    return response[:received], nil
}
