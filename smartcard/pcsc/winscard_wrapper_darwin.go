// +build darwin

package pcsc

//#cgo CFLAGS: -framework PCSC
//#cgo LDFLAGS: -framework PCSC
//#include <PCSC/winscard.h>
//#include <PCSC/wintypes.h>
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"
)

type readerState struct {
	reader       uintptr
	userData     uintptr
	currentState uint32
	eventState   uint32
	atrLen       uint32
	atr          [_MAX_ATR_SIZE]byte
}

type ReaderState struct {
	Reader       string
	UserData     uintptr
	CurrentState uint32
	EventState   uint32
	AtrLen       uint32
	Atr          [_MAX_ATR_SIZE]byte
}

type WinscardWrapper struct {
	t0PCI uintptr
	t1PCI uintptr
}

var theWrapper *WinscardWrapper = nil

func Winscard() (*WinscardWrapper, error) {
	if theWrapper != nil {
		return theWrapper, nil
	}
	t0 := C.g_rgSCardT0Pci
	t1 := C.g_rgSCardT1Pci
	winscard := &WinscardWrapper{
		t0PCI: uintptr(unsafe.Pointer(&t0)),
		t1PCI: uintptr(unsafe.Pointer(&t1)),
	}
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

func (ww *WinscardWrapper) EstablishContext() (uintptr, error) {

	var hContext C.SCARDCONTEXT
	rv := C.SCardEstablishContext(C.SCARD_SCOPE_SYSTEM, nil, nil, &hContext)

	if rv != SCARD_S_SUCCESS {
		return 0, fmt.Errorf("can't establish context: %s",
			errorString(uint32(rv)))
	}
	return uintptr(hContext), nil
}

func (ww *WinscardWrapper) ReleaseContext(ctx uintptr) error {
	rv := C.SCardReleaseContext(C.int(ctx))

	if rv != SCARD_S_SUCCESS {
		return fmt.Errorf("can't release context: %s",
			errorString(uint32(rv)))
	}
	return nil
}

func (ww *WinscardWrapper) ListReaders(ctx uintptr) ([]string, error) {
	var bufferSize C.DWORD
	var rv C.LONG
	readers := make([]string, 0, 3)
	rv = C.SCardListReaders(C.int(ctx), nil, nil, &bufferSize)
	if rv != SCARD_S_SUCCESS {
		if uint64(rv) == SCARD_E_NO_READERS_AVAILABLE {
			return readers, nil
		}
		return nil, fmt.Errorf("can't list readers: %s",
			errorString(uint32(rv)))
	}
	buffer := make([]byte, bufferSize)
	rv = C.SCardListReaders(C.int(ctx), nil, (*C.char)(unsafe.Pointer(&buffer[0])), &bufferSize)

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
	rv := C.SCardGetStatusChange(C.int(ctx), C.uint(timeout),
		(C.LPSCARD_READERSTATE_A)(unsafe.Pointer(&_states[0])), C.uint(len(_states)))
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
	if err != nil {
		return nil, err
	}
	count := len(readerNames)
	if count == 0 {
		return nil, nil
	}
	states := make([]ReaderState, 1)
	for i := 0; i < len(readerNames); i++ {
		states[i].Reader = readerNames[i]
		states[i].CurrentState = currentState
	}
	err = ww.GetStatusChange(ctx, SCARD_INFINITE, states)
	if err != nil {
		return nil, err
	}
	return states, nil
}

func (ww *WinscardWrapper) CardConnect(ctx uintptr, reader string) (
	uintptr, uintptr, error) {
	var card C.int
	var activeProtocol C.uint
	rv := C.SCardConnect(C.int(ctx), (*C.char)(unsafe.Pointer(&ww.stringToBytes(reader)[0])), C.SCARD_SHARE_SHARED,
		C.SCARD_PROTOCOL_ANY, &card, &activeProtocol)
	if rv != SCARD_S_SUCCESS {
		return 0, 0, fmt.Errorf("can't connect to card: %s",
			errorString(uint32(rv)))
	}
	return uintptr(card), uintptr(activeProtocol), nil
}

func (ww *WinscardWrapper) CardDisconnect(card uintptr) error {
	rv := C.SCardDisconnect(C.int(card), SCARD_RESET_CARD)
	if rv != SCARD_S_SUCCESS {
		return fmt.Errorf("can't disconnect from card: %s",
			errorString(uint32(rv)))
	}
	return nil
}

func (ww *WinscardWrapper) Transmit(card uintptr, sendPCI uintptr,
	sendBuffer []byte, recvBuffer []byte) (uint32, error) {
	received := C.uint(len(recvBuffer))
	rv := C.SCardTransmit(C.int(card),
		(*C.SCARD_IO_REQUEST)(unsafe.Pointer(sendPCI)),
		(*C.uchar)(unsafe.Pointer(&sendBuffer[0])),
		C.uint(len(sendBuffer)),
		nil,
		(*C.uchar)(unsafe.Pointer(&recvBuffer[0])),
		&received)

	if rv != SCARD_S_SUCCESS {
		return 0, fmt.Errorf("transmission failed: %s",
			errorString(uint32(rv)))
	}

	return uint32(received), nil
}

func (ww WinscardWrapper) GetAttrib(card uintptr, attr uint32) ([]byte, error) {
	var size C.uint

	rv := C.SCardGetAttrib(C.int(card), C.uint(attr),
		nil, &size)
	if rv != SCARD_S_SUCCESS {
		return nil, fmt.Errorf("can't get attribute : %s",
			errorString(uint32(rv)))
	}
	buffer := make([]byte, size)
	rv = C.SCardGetAttrib(C.int(card),
		C.uint(attr),
		(*C.uchar)(unsafe.Pointer(&buffer[0])),
		&size)
	if rv != SCARD_S_SUCCESS {
		return nil, fmt.Errorf("can't get attribute : %s",
			errorString(uint32(rv)))
	}
	return buffer[:size], nil
}
