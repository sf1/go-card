package pcsclite

func ListReaders() ([]*Reader, error) {
    return listReaders(false)
}

func ListReadersWithCard() ([]*Reader, error) {
    return listReaders(true)
}

func listReaders(withCard bool) ([]*Reader, error) {
    client, err := PCSCDConnect()
    if err != nil { return nil, err}
    defer client.Close()
    readerInfos, err := client.ListReaders()
    if err != nil { return nil, err }
    result := make([]*Reader, 0, len(readerInfos))
    for i := 0; i < len(readerInfos); i++ {
        if withCard {
            if readerInfos[i].IsCardPresent() {
                result = append(result, &Reader{*readerInfos[i]})
            }
        } else {
            result = append(result, &Reader{*readerInfos[i]})
        }
    }
    return result, nil
}

func WaitForCardPresent() (*Reader, error) {
    var reader *Reader
    client, err := PCSCDConnect()
    if err != nil { return nil, err}
    defer client.Close()
    for reader == nil {
        count, err := client.SyncReaderStates()
        if err != nil { return nil, err}
        for i := uint32(0); i < count; i++ {
            if client.readerStates[i].IsCardPresent() {
                reader = &Reader{client.readerStates[i]}
                break
            }
        }
        client.WaitReaderStateChange()
    }
    return reader, nil
}
