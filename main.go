package main

import (
    "fmt"
    "github.com/viscory/reciprocus/blockchain"
    "strconv"
)

func main() {
    chain := blockchain.InitBlockChain()
    chain.AddBlock("First After Gen")
    chain.AddBlock("Second After Gen")
    chain.AddBlock("Third After Gen")

    for _,block := range chain.Blocks {
        fmt.Printf("Previous Hash: %x\n", block.PrevHash)
        fmt.Printf("Data in Block: %s\n", block.Data)
        fmt.Printf("Hash: %x\n", block.Hash)

        pow:= blockchain.NewProof(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
        fmt.Println()
    }
}
