package main

import (
    "fmt"
    "github.com/viscory/reciprocus/blockchain"
    "strconv"
)

func main() {
    chain := InitBlockChain()
    chain.AddBlock("First")
    chain.AddBlock("Second")
    chain.AddBlock("Third")

    for _,block := range chain.blocks {
        fmt.Printf("Previous Hash: %x\n", block.PrevHash)
        fmt.Printf("Data in Block: %s\n", block.Data)
        fmt.Printf("Hash: %x\n", block.Hash)
        fmt.Println()

        pow:= blockchain.NewProof(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
        fmt.Println()
    }
}
