// +build !windows

package pcsc

import (
    "net"
    "unsafe"
    "errors"
    "bytes"
    "fmt"
)

type rxHeader struct {
    size uint32
    command uint32
}

type versionStruct struct {
    major int32
    minor int32
    rv uint32
}

type establishStruct struct {
    scope uint32
    context uint32
    rv uint32
}

type releaseStruct struct {
    context uint32
    rv uint32
}

type connectStruct struct {
    context uint32
    readerName [_MAX_READERNAME]byte
    shareMode uint32
    preferredProtocols uint32
    card int32
    activeProtocol uint32
    rv uint32
}

type disconnectStruct struct {
    card int32
    disposition uint32
    rv uint32
}

type transmitStruct struct {
    card int32
    sendPciProtocol uint32
    sendPciLength uint32
    sendLength uint32
    recvPciProtocol uint32
    recvPciLength uint32
    recvLength uint32
    rv uint32
}

type waitReaderStateChangeStruct struct {
    timeOutMs uint32
    rv uint32
}

type ReaderInfo struct {
    ReaderName [_MAX_READERNAME]byte
    EventCounter uint32
    ReaderState uint32
    ReaderSharing int32
    CardAtr [_MAX_ATR_SIZE] byte
    CardAtrLength uint32
    CardProtocol uint32
}

func (ri *ReaderInfo) Name() string {
    n := bytes.IndexByte(ri.ReaderName[:], 0)
    return string(ri.ReaderName[:n])
}

func (ri *ReaderInfo) IsCardPresent() bool {
    present := uint32(_SCARD_POWERED | _SCARD_PRESENT)
    return (ri.ReaderState & present) == present
}

func (ri *ReaderInfo) String() string {
    var buffer bytes.Buffer
    buffer.WriteString(ri.Name())
    buffer.WriteString("\n")
    buffer.WriteString(fmt.Sprintf("- Event Counter:  %d\n", ri.EventCounter))
    buffer.WriteString(fmt.Sprintf("- Reader State:   %x (", ri.ReaderState))
    if (ri.ReaderState & _SCARD_SPECIFIC) != 0 {
        buffer.WriteString(" SPECFIC")
    }
    if (ri.ReaderState & _SCARD_NEGOTIABLE) != 0 {
        buffer.WriteString(" NEGOTIABLE")
    }
    if (ri.ReaderState & _SCARD_POWERED) != 0 {
        buffer.WriteString(" POWERED")
    }
    if (ri.ReaderState & _SCARD_SWALLOWED) != 0 {
        buffer.WriteString(" SWALLOWED")
    }
    if (ri.ReaderState & _SCARD_PRESENT) != 0 {
        buffer.WriteString(" PRESENT")
    }
    if (ri.ReaderState & _SCARD_ABSENT) != 0 {
        buffer.WriteString(" ABSENT")
    }
    if (ri.ReaderState & _SCARD_UNKNOWN) != 0 {
        buffer.WriteString(" UNKOWN")
    }
    buffer.WriteString(" )\n")
    buffer.WriteString(fmt.Sprintf("- Reader Sharing: %d\n", ri.ReaderSharing))
    buffer.WriteString(fmt.Sprintf("- Card ATR Len:   %d\n", ri.CardAtrLength))
    buffer.WriteString("- Card ATR:       ")
    for y := uint32(0); y < ri.CardAtrLength; y++ {
        buffer.WriteString(fmt.Sprintf("%02x", ri.CardAtr[y]))
    }
    buffer.WriteString(fmt.Sprintf("\n- Card Protocol:  %08x\n", ri.CardProtocol))
    return buffer.String()
}

type ReaderStateArray [_PCSCLITE_MAX_READERS_CONTEXTS]ReaderInfo

type PCSCLiteClient struct {
    connection net.Conn
    readerStates ReaderStateArray
    readerCount uint32
}

func PCSCLiteConnect() (*PCSCLiteClient, error) {
    var err error
    client := &PCSCLiteClient{}
    client.connection, err = net.Dial("unix","/var/run/pcscd/pcscd.comm")
    if err != nil { return nil, errors.New("Can't connect to PCSCD") }
    version := versionStruct{_PROTOCOL_VERSION_MAJOR, _PROTOCOL_VERSION_MINOR, 0}
    ptr1 := (*[unsafe.Sizeof(version)]byte)(unsafe.Pointer(&version))
    err = client.ExchangeMessage(_CMD_VERSION, ptr1[:])
    if err != nil { return nil, err }
    if version.rv != _SCARD_S_SUCCESS {
        return nil, errors.New("Protocol version mismatch")
    }
    return client, nil
}

func (client* PCSCLiteClient) ReaderStates() ReaderStateArray {
    return client.readerStates
}

func (client *PCSCLiteClient) Close() {
    client.connection.Close()
}

func (client *PCSCLiteClient) SendHeader(command uint32, msgLen uint32) error {
    header := rxHeader{msgLen, command}
    headerPtr := (*[unsafe.Sizeof(header)]byte)(unsafe.Pointer(&header))
    _, err := client.connection.Write(headerPtr[:])
    return err
}

func (client *PCSCLiteClient) ExchangeMessage(command uint32, msg []byte) error {
    err := client.SendHeader(command, uint32(len(msg)))
    if err != nil { return err }
    _, err = client.connection.Write(msg)
    if err != nil { return err }
    _, err = client.connection.Read(msg)
    return err
}

func (client *PCSCLiteClient) Read(data []byte) (int, error) {
    return client.connection.Read(data)
}

func (client *PCSCLiteClient) Write(data []byte) (int, error) {
    return client.connection.Write(data)
}

func (client *PCSCLiteClient) EstablishContext() (uint32, error) {
    estruct := establishStruct{_CARD_SCOPE_SYSTEM, 0, 0}
    ptr := (*[unsafe.Sizeof(estruct)]byte)(unsafe.Pointer(&estruct))
    err := client.ExchangeMessage(_SCARD_ESTABLISH_CONTEXT, ptr[:])
    if err != nil { return 0, err }
    if estruct.rv != _SCARD_S_SUCCESS {
        return 0, fmt.Errorf("Can't establish context: %08x", estruct.rv)
    }
    return estruct.context, nil
}

func (client *PCSCLiteClient) ReleaseContext(context uint32) error {
    rstruct := releaseStruct{context, 0}
    ptr := (*[unsafe.Sizeof(rstruct)]byte)(unsafe.Pointer(&rstruct))
    err := client.ExchangeMessage(_SCARD_RELEASE_CONTEXT, ptr[:])
    if err != nil { return err }
    if rstruct.rv != _SCARD_S_SUCCESS {
        return fmt.Errorf("Can't release context: %08x", rstruct.rv)
    }
    return nil
}

func (client *PCSCLiteClient) SyncReaderStates() (
    uint32, error) {
    var count uint32
    ptr := (*[unsafe.Sizeof(client.readerStates)]byte)(
        unsafe.Pointer(&client.readerStates))
    err := client.SendHeader(_CMD_GET_READERS_STATE, 0)
    _, err = client.Read(ptr[:])
    if err != nil { return count, err }
    for count = 0; count < _PCSCLITE_MAX_READERS_CONTEXTS; count++ {
        ri := client.readerStates[count]
        if ri.ReaderName[0] == 0 {
            break
        }
    }
    client.readerCount = count
    return count, nil
}

func (client *PCSCLiteClient) ListReaders() ([]*ReaderInfo, error) {
    client.SyncReaderStates()
    readers := make([]*ReaderInfo, client.readerCount)
    for i := uint32(0); i < client.readerCount; i++ {
        readers[i] = &client.readerStates[i]
    }
    return readers, nil
}

func (client *PCSCLiteClient) CardConnect(context uint32, readerName string) (
    int32, uint32, error) {
    cstruct := connectStruct{}
    cstruct.context = context
    readerBytes := ([]byte)(readerName)
    limit := len(readerBytes)
    if limit > _MAX_READERNAME { limit = _MAX_READERNAME }
    for i := 0; i < limit; i++ {
        cstruct.readerName[i] = readerBytes[i]
    }
    cstruct.shareMode = _SCARD_SHARE_SHARED
    cstruct.preferredProtocols = _SCARD_PROTOCOL_ANY
    ptr := (*[unsafe.Sizeof(cstruct)]byte)(unsafe.Pointer(&cstruct))
    err := client.ExchangeMessage(_SCARD_CONNECT, ptr[:])
    if err != nil { return 0, 0, err }
    if cstruct.rv != _SCARD_S_SUCCESS {
        return 0, 0, fmt.Errorf("Cant connect to card: %08x", cstruct.rv)
    }
    return cstruct.card, cstruct.activeProtocol, nil
}

func (client *PCSCLiteClient) CardDisconnect(card int32) error {
    dstruct := disconnectStruct{
        card,
        _SCARD_RESET_CARD,
        0,
    }
    ptr := (*[unsafe.Sizeof(dstruct)]byte)(unsafe.Pointer(&dstruct))
    err := client.ExchangeMessage(_SCARD_DISCONNECT, ptr[:])
    if err != nil { return err }
    if dstruct.rv != _SCARD_S_SUCCESS {
        return fmt.Errorf("Cant disconnect from card: %08x", dstruct.rv)
    }
    return nil
}

func (client *PCSCLiteClient) Transmit(card int32, protocol uint32,
    sendBuffer []byte, recvBuffer []byte) (uint32, error) {
    tstruct := transmitStruct{}
    tstruct.card = card
    tstruct.sendLength = uint32(len(sendBuffer))
    tstruct.sendPciProtocol = protocol
    tstruct.sendPciLength = 8
    tstruct.recvLength = uint32(len(recvBuffer))
    tstruct.recvPciProtocol = _SCARD_PROTOCOL_ANY
    tstruct.recvPciLength = 8
    tsBytes := (*[unsafe.Sizeof(tstruct)]byte)(unsafe.Pointer(&tstruct))[:]
    err := client.SendHeader(_SCARD_TRANSMIT, uint32(len(tsBytes)))
    if err != nil { return 0, err }
    _, err = client.connection.Write(tsBytes)
    if err != nil { return 0, err }
    _, err = client.connection.Write(sendBuffer)
    if err != nil { return 0, err }
    _, err = client.connection.Read(tsBytes)
    if err != nil { return 0, err }
    if tstruct.rv != _SCARD_S_SUCCESS {
        return 0, fmt.Errorf("Transmission failed: %08x", tstruct.rv)
    }
    _, err = client.connection.Read(recvBuffer)
    if err != nil { return 0, err }
    return tstruct.recvLength, nil
}

func (client *PCSCLiteClient) WaitReaderStateChange() error {
    wrstruct := waitReaderStateChangeStruct{ uint32(60000), 0 }
    ptr := (*[unsafe.Sizeof(wrstruct)]byte)(unsafe.Pointer(&wrstruct))
    err := client.ExchangeMessage(_CMD_WAIT_READER_STATE_CHANGE, ptr[:])
    if err != nil { return err }
    if wrstruct.rv == _SCARD_E_TIMEOUT {
        client.ExchangeMessage(_CMD_STOP_WAITING_READER_STATE_CHANGE, ptr[:])
        if err != nil { return err }
    }
    if wrstruct.rv != _SCARD_S_SUCCESS {
        return fmt.Errorf("Wait failed: %08x", wrstruct.rv)
    }
    return nil
}
