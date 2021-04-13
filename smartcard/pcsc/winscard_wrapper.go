// +build windows

package pcsc

import (
    "fmt"
    "bytes"
    "unsafe"
    "syscall"
)

type readerState struct {
    reader uintptr
    userData uintptr
    currentState uint32
    eventState uint32
    atrLen uint32
    atr [_MAX_ATR_SIZE]byte
}

type ReaderState struct {
    Reader string
    UserData uintptr
    CurrentState uint32
    EventState uint32
    AtrLen uint32
    Atr [_MAX_ATR_SIZE]byte
}

type WinscardWrapper struct {
    establishContext *syscall.LazyProc
    releaseContext *syscall.LazyProc
    listReaders *syscall.LazyProc
    cardConnect *syscall.LazyProc
    cardDisconnect *syscall.LazyProc
    transmit *syscall.LazyProc
    getStatusChange *syscall.LazyProc
    getAttrib *syscall.LazyProc
    t0PCI uintptr
    t1PCI uintptr
}

var theWrapper *WinscardWrapper = nil

func Winscard() (*WinscardWrapper, error) {
    if theWrapper != nil {
        return theWrapper, nil
    }
    dll := syscall.NewLazyDLL("winscard.dll")
    winscard := &WinscardWrapper{}
    theWrapper = winscard
    winscard.establishContext = dll.NewProc("SCardEstablishContext")
    err := winscard.establishContext.Find()
    if err != nil { return nil, err }
    winscard.releaseContext = dll.NewProc("SCardReleaseContext")
    winscard.listReaders = dll.NewProc("SCardListReadersA")
    winscard.cardConnect = dll.NewProc("SCardConnectA")
    winscard.cardDisconnect = dll.NewProc("SCardDisconnect")
    winscard.transmit = dll.NewProc("SCardTransmit")
    winscard.getStatusChange = dll.NewProc("SCardGetStatusChangeA")
    winscard.getAttrib = dll.NewProc("SCardGetAttrib")
    t0 := dll.NewProc("g_rgSCardT0Pci")
    t1 := dll.NewProc("g_rgSCardT1Pci")
    if t0.Find() != nil || t1.Find() != nil {
        fmt.Errorf("pci structures not found")
    }
    winscard.t0PCI = t0.Addr()
    winscard.t1PCI = t1.Addr()
    return winscard, nil
}

func (ww *WinscardWrapper) T0PCI() uintptr {
    return ww.t0PCI
}

func (ww *WinscardWrapper) T1PCI() uintptr {
    return ww.t1PCI
}

func (ww *WinscardWrapper) stringToBytes(str string) []byte {
    var buffer bytes.Buffer
    buffer.WriteString(str)
    buffer.WriteByte(0)
    return buffer.Bytes()
}

func (ww *WinscardWrapper) EstablishContext(scope ...uint32) (uintptr, error) {
    var ctx uintptr
    scp := uint32(CARD_SCOPE_SYSTEM)
    if len(scope) > 0 {
        scp = scope[0]
    }
    rv, _, _ := ww.establishContext.Call(uintptr(scp), uintptr(0),
        uintptr(0), uintptr(unsafe.Pointer(&ctx)))
    if rv != SCARD_S_SUCCESS {
        return 0, fmt.Errorf("can't establish context: %s",
            errorString(uint32(rv)))
    }
    return ctx, nil
}

func (ww *WinscardWrapper) ReleaseContext(ctx uintptr) error {
    rv, _, _ := ww.releaseContext.Call(uintptr(ctx))
    if rv != SCARD_S_SUCCESS {
        return fmt.Errorf("can't release context: %s",
            errorString(uint32(rv)))
    }
    return nil
}

func (ww *WinscardWrapper) ListReaders(ctx uintptr) ([]string, error) {
    var bufferSize uintptr
    readers := make([]string, 0, 3)
    rv, _, _ := ww.listReaders.Call(ctx, 0, 0,
        uintptr(unsafe.Pointer(&bufferSize)))
    if rv != SCARD_S_SUCCESS {
        if rv == SCARD_E_NO_READERS_AVAILABLE {
            return readers, nil
        }
        return nil, fmt.Errorf("can't list readers: %s",
            errorString(uint32(rv)))
    }
    buffer := make([]byte, bufferSize)
    rv, _, _ = ww.listReaders.Call(ctx, 0,
        uintptr(unsafe.Pointer(&buffer[0])),
        uintptr(unsafe.Pointer(&bufferSize)))
    if rv != SCARD_S_SUCCESS {
        return nil, fmt.Errorf("can't list readers: %s",
            errorString(uint32(rv)))
    }
    n := bytes.IndexByte(buffer, 0)
    for n != 0 {
        readers = append(readers, string(buffer[:n]))
        buffer = buffer[n+1:]
        n = bytes.IndexByte(buffer, 0)
    }
    return readers, nil
}

func (ww *WinscardWrapper) GetStatusChange(ctx uintptr, timeout uint32,
    states []ReaderState) error {
    _states := make([]readerState, len(states))
    for i := 0; i < len(states); i++ {
        var buffer bytes.Buffer
        buffer.WriteString(states[i].Reader)
        buffer.WriteByte(0)
        _states[i].reader = uintptr(unsafe.Pointer(&buffer.Bytes()[0]))
        _states[i].currentState = states[i].CurrentState
    }
    rv, _, _ := ww.getStatusChange.Call(ctx, uintptr(timeout),
        uintptr(unsafe.Pointer(&_states[0])), uintptr(len(_states)))
    if rv != SCARD_S_SUCCESS {
        return fmt.Errorf("get status change failed: %s",
            errorString(uint32(rv)))
    }
    for i := 0; i < len(states); i++ {
        states[i].UserData = _states[i].userData
        states[i].EventState = _states[i].eventState
        states[i].AtrLen = _states[i].atrLen
        states[i].Atr = _states[i].atr
    }
    return nil
}

func (ww *WinscardWrapper) GetStatusChangeAll(ctx uintptr, timeout uint32,
    currentState uint32) ([]ReaderState, error) {
    readerNames, err := ww.ListReaders(ctx)
    if err != nil { return nil, err }
    count := len(readerNames)
    if count == 0 {
        return nil, nil
    }
    states := make([]ReaderState, len(readerNames))
    for i := 0; i < len(readerNames); i++ {
        states[i].Reader = readerNames[i]
        states[i].CurrentState = currentState
    }
    err = ww.GetStatusChange(ctx, SCARD_INFINITE, states)
    if err != nil { return nil, err }
    return states, nil
}

func (ww *WinscardWrapper) CardConnect(ctx uintptr, reader string) (
    uintptr, uintptr, error) {
    var card, activeProtocol uintptr
    rv, _, _ := ww.cardConnect.Call(
        ctx,
        uintptr(unsafe.Pointer(unsafe.Pointer(&ww.stringToBytes(reader)[0]))),
        uintptr(SCARD_SHARE_SHARED), uintptr(SCARD_PROTOCOL_ANY),
        uintptr(unsafe.Pointer(&card)),
        uintptr(unsafe.Pointer(&activeProtocol)),
    )
    if rv != SCARD_S_SUCCESS {
        return 0, 0, fmt.Errorf("can't connect to card: %s",
            errorString(uint32(rv)))
    }
    return card, activeProtocol, nil
}

func (ww *WinscardWrapper) CardDisconnect(card uintptr) error {
    rv, _, _ := ww.cardDisconnect.Call(card, uintptr(SCARD_RESET_CARD))
    if rv != SCARD_S_SUCCESS {
        return fmt.Errorf("can't disconnect from card: %s",
            errorString(uint32(rv)))
    }
    return nil
}

func (ww *WinscardWrapper) Transmit(card uintptr, sendPCI uintptr,
    sendBuffer []byte, recvBuffer []byte) (uint32, error) {
        received := uint32(len(recvBuffer))
	rv, _, _ := ww.transmit.Call(card,
		sendPCI,
		uintptr(unsafe.Pointer(&sendBuffer[0])),
		uintptr(len(sendBuffer)),
		0,
		uintptr(unsafe.Pointer(&recvBuffer[0])),
		uintptr(unsafe.Pointer(&received)))
        if rv != SCARD_S_SUCCESS {
            return 0, fmt.Errorf("transmission failed: %s",
                errorString(uint32(rv)))
        }
        return received, nil
}

func (ww WinscardWrapper) GetAttrib(card uintptr, attr uint32) ([]byte, error) {
    var size uintptr
    rv, _, _ := ww.getAttrib.Call(
        card, uintptr(attr), 0,
        uintptr(unsafe.Pointer(&size)),
    )
    if rv != SCARD_S_SUCCESS {
        return nil, fmt.Errorf("can't get attribute : %s",
            errorString(uint32(rv)))
    }
    buffer := make([]byte, size)
    rv, _, _ = ww.getAttrib.Call(
        card, uintptr(attr),
        uintptr(unsafe.Pointer(&buffer[0])),
        uintptr(unsafe.Pointer(&size)),
    )
    if rv != SCARD_S_SUCCESS {
        return nil, fmt.Errorf("can't get attribute : %s",
            errorString(uint32(rv)))
    }
    return buffer[:size], nil
}
