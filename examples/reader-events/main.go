package main

import (
    "fmt"
    "github.com/sf1/go-sc/smartcard"
)

func printReaderStates() {
    readers, err := smartcard.ListReaders()
    if err != nil { panic(err) }
    for _, reader := range readers {
        fmt.Printf("%s\n", reader.Name())
        fmt.Printf("- Card present: %t\n\n", reader.IsCardPresent())
    }
}

func main() {
    fmt.Println("\nReader List")
    fmt.Println("-----------\n")
    printReaderStates()
    fmt.Println("Reader Events")
    fmt.Println("-------------\n")
    fmt.Println("Waiting...\n")
}
