// +build windows

package smartcard

import (
    "fmt"
    "unsafe"
    "syscall"
    "bytes"
)

var (
    winscardDLL = syscall.NewLazyDLL("winscard.dll")
    scardEstablishContext = winscardDLL.NewProc("SCardEstablishContext")
    scardReleaseContext = winscardDLL.NewProc("SCardReleaseContext")
    scardListReaders = winscardDLL.NewProc("SCardListReadersA")
    scardConnect = winscardDLL.NewProc("SCardConnect")
    scardDisconnect = winscardDLL.NewProc("SCardDisconnect")
    scardTransmit = winscardDLL.NewProc("SCardTransmit")
    scardGetStatusChange = winscardDLL.NewProc("SCardGetStatusChangeA")
)

func EstablishContext() (Context, error) {
    var ctxID uintptr
    err := scardEstablishContext.Find()
    if err != nil { return nil, err }
    rv, _, _ := scardEstablishContext.Call(CARD_SCOPE_SYSTEM, uintptr(0),
        uintptr(0), uintptr(unsafe.Pointer(&ctxID)))
    if rv != SCARD_S_SUCCESS {
        return nil, fmt.Errorf("Can't establish context: %08x", rv)
    }
    return &WinscardContext{ctxID}, nil
}

type WinscardContext struct {
    ctxID uintptr
}

func (ctx *WinscardContext) Release() error {
    err := scardReleaseContext.Find()
    if err != nil { return err }
    rv, _, _ := scardReleaseContext.Call(uintptr(ctx.ctxID))
    if rv != SCARD_S_SUCCESS {
        return fmt.Errorf("Can't release context: %08x", rv)
    }
    return nil
}

func (ctx *WinscardContext) ListReaders() ([]Reader, error) {
    var bufferSize uintptr
    err := scardListReaders.Find()
    if err != nil { return nil, err }
    rv, _, _ := scardListReaders.Call(ctx.ctxID, 0, 0,
        uintptr(unsafe.Pointer(&bufferSize)))
    if rv != SCARD_S_SUCCESS {
        return nil, fmt.Errorf("Can't list readers: %08x", rv)
    }
    buffer := make([]byte, bufferSize)
    rv, _, _ = scardListReaders.Call(ctx.ctxID, 0,
        uintptr(unsafe.Pointer(&buffer[0])),
        uintptr(unsafe.Pointer(&bufferSize)))
    if rv != SCARD_S_SUCCESS {
        return nil, fmt.Errorf("Can't list readers: %08x", rv)
    }
    readers := make([]Reader, 0, 3)
    n := bytes.IndexByte(buffer, 0)
    for n != 0 {
        readers = append(readers,
            &WinscardReader{ctx.ctxID, buffer[:n+1]})
        buffer = buffer[n+1:]
        n = bytes.IndexByte(buffer, 0)
    }
    return readers, nil
}

func (ctx *WinscardContext) ListReadersWithCard() ([]Reader, error) {
    return nil, nil
}

func (ctx *WinscardContext) WaitForCardPresent() (Reader, error) {
    return nil, nil
}

type readerState struct {
    szReader uintptr
    pvUserData uintptr
    dwCurrentState uint32
    dwEventState uint32
    cbAtr uint32
    gbAtr [MAX_ATR_SIZE]byte
}

type WinscardReader struct {
    ctxID uintptr
    name []byte
}

func (r *WinscardReader) Name() string {
    return string(r.name[:len(r.name)-1])
}

func (r *WinscardReader) IsCardPresent() bool {
    err := scardGetStatusChange.Find()
    if err != nil { return false }
    states := make([]readerState, 1)
    states[0].szReader = uintptr(unsafe.Pointer(&r.name[0]))
    states[0].dwCurrentState = SCARD_STATE_UNAWARE
    rv, _, _ := scardGetStatusChange.Call(r.ctxID, uintptr(SCARD_INFINITE),
        uintptr(unsafe.Pointer(&states[0])), uintptr(len(states)))
    if rv != SCARD_S_SUCCESS {
        return false
    }
    if states[0].dwEventState & SCARD_STATE_MUTE != 0 {
        return false
    }
    return states[0].dwEventState & SCARD_STATE_PRESENT != 0
}

func (r *WinscardReader) Connect() (Card, error) {
    return nil, nil
}
