// +build windows

package smartcard

import (
    "fmt"
    "time"
    "github.com/sf1/go-card/smartcard/pcsc"
)

// A smart card context is required to access readers and cards.
type Context struct {
    ctxID uintptr
    winscard *pcsc.WinscardWrapper
}

// Establish smart card context.
// This should be the first function to be called.
func EstablishContext(scope uint32) (*Context, error) {
    winscard, err := pcsc.Winscard()
    if err != nil {
        return nil, err
    }
    ctxID, err := winscard.EstablishContext(scope)
    if err != nil { return nil, err }
    return &Context{ctxID, winscard}, nil
}

// Release resources associated with smart card context.
func (ctx *Context) Release() error {
    return ctx.winscard.ReleaseContext(ctx.ctxID)
}

// List all smart card readers.
func (ctx *Context) ListReaders() ([]*Reader, error) {
    readerNames, err := ctx.winscard.ListReaders(ctx.ctxID)
    if err != nil { return nil, err }
    readers := make([]*Reader, len(readerNames))
    for i := 0; i < len(readerNames); i++ {
        readers[i] = &Reader{context: ctx, name: readerNames[i]}
    }
    return readers, nil
}

// List smart card readers with inserted cards.
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
            readers = append(readers, &Reader{context: ctx, name: state.Reader})
        }
    }
    return readers, nil
}

// Block until a smart card is inserted into any reader.
// Returns immediately if card already present.
func (ctx *Context) WaitForCardPresent() (*Reader, error) {
    var err error
    var reader *Reader = nil
    var states []pcsc.ReaderState
    for reader == nil {
        states, err = ctx.winscard.GetStatusChangeAll(
            ctx.ctxID, pcsc.SCARD_INFINITE, pcsc.SCARD_STATE_UNAWARE)
        if err != nil { return nil, err }
        if states == nil {
            time.Sleep(500*time.Millisecond)
            continue
        }
        for _, state := range states {
            state.CurrentState = state.EventState
            if state.EventState & pcsc.SCARD_STATE_MUTE != 0 {
                continue
            }
            if state.EventState & pcsc.SCARD_STATE_PRESENT != 0 {
                reader = &Reader{context: ctx, name: state.Reader}
                break
            }
        }
        if reader == nil {
            err = ctx.winscard.GetStatusChange(
                ctx.ctxID, pcsc.SCARD_INFINITE, states,
            )
            if err != nil { return nil, err }
            time.Sleep(500*time.Millisecond)
        }
    }
    return reader, nil
}

// Smart card reader. 
// Note that physical card readers with slots for multiple cards are
// represented by one Reader instance per slot.
type Reader struct {
    context *Context
    name string
}

// Return name of card reader.
func (r *Reader) Name() string {
    return r.name
}

// Check if card is present.
func (r *Reader) IsCardPresent() bool {
    states := make([]pcsc.ReaderState, 1)
    states[0].Reader = r.name
    states[0].CurrentState = pcsc.SCARD_STATE_UNAWARE
    err := r.context.winscard.GetStatusChange(r.context.ctxID,
        pcsc.SCARD_INFINITE, states)
    if err != nil {
        return false
    }
    if states[0].EventState & pcsc.SCARD_STATE_MUTE != 0 {
        return false
    }
    return states[0].EventState & pcsc.SCARD_STATE_PRESENT != 0
}

// Wait until card removed
func (r *Reader) WaitUntilCardRemoved() {
    states := make([]pcsc.ReaderState, 1)
    for {
        states[0].Reader = r.name
        states[0].CurrentState = pcsc.SCARD_STATE_UNAWARE
        err := r.context.winscard.GetStatusChange(r.context.ctxID,
            pcsc.SCARD_INFINITE, states)
        if err != nil {
            return
        }
        if states[0].EventState & pcsc.SCARD_STATE_MUTE != 0 ||
            states[0].EventState & pcsc.SCARD_STATE_PRESENT == 0 {
                return
        }
        time.Sleep(500*time.Millisecond)
    }
}


// Connect to card.
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
    return &Card{
        context:r.context, 
        cardID: cardID,
        sendPCI: pci, 
    }, nil
}

// Smart card.
type Card struct {
    context *Context
    cardID uintptr
    sendPCI uintptr
    atr ATR
}

// Disconnect from card.
func (c *Card) Disconnect() error {
    err := c.context.winscard.CardDisconnect(c.cardID)
    if err != nil { return err }
    return nil
}

// Return card ATR (answer to reset).
func (c *Card) ATR() ATR {
    var err error
    if c.atr != nil { return c.atr }
    c.atr, err = c.context.winscard.GetAttrib(c.cardID, pcsc.SCARD_ATTR_ATR_STRING)
    if err != nil { return nil }
    return c.atr
}

// Trasmit bytes to card and return response.
func (c *Card) Transmit(command []byte) ([]byte, error) {
    response := make([]byte, 258)
    received, err := c.context.winscard.Transmit(c.cardID, c.sendPCI,
        command, response)
    if err != nil { return nil, err }
    return response[:received], nil
}
