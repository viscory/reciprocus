package cli

import (
    "fmt"
    "flag"
    "os"
    "runtime"
    "github.com/viscory/reciprocus/blockchain"
    "github.com/viscory/reciprocus/wallet"
    "strconv"
)

type CommandLine struct {
    blockchain *blockchain.BlockChain
}

func (cli *CommandLine) printUsage() {
    fmt.Println("Usage:")
    fmt.Println(" printchain - Prints the blocks in the chain")
    fmt.Println(" getbalance -adress ADDRESS - get the balance for address")
    fmt.Println(" send -from FROM -to TO -amount AMOUNT - send AMOUNT to TO from FROM")
    fmt.Println(" createblockchain -address ADDRESS create(mine) a blockchain")
    fmt.Println(" createwallet - create new wallet")
    fmt.Println(" getallwallets - lists all wallets inside wallet file")
}

func (cli *CommandLine) validateArgs() {
    if len(os.Args) < 2 {
        cli.printUsage()
        runtime.Goexit()
    }
}

func (cli *CommandLine) printChain() {
    chain := blockchain.ContinueBlockChain("")
    defer chain.Database.Close()
    iter := chain.Iterator()

    for {
        block := iter.Next()

        fmt.Printf("Previous Hash: %x\n", block.PrevHash)
        fmt.Printf("Hash: %x\n", block.Hash)
        pow:= blockchain.NewProof(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
        fmt.Println()

        if len(block.PrevHash) == 0 {
            break
        }
    }
}

func (cli *CommandLine) createBlockChain(address string) {
    chain := blockchain.InitBlockChain(address)
    chain.Database.Close()
    fmt.Println("Created Blockchain")
}

func (cli *CommandLine) getBalance(address string) {
    chain := blockchain.ContinueBlockChain(address)
    defer chain.Database.Close()

    balance := 0
    UTXOs := chain.FindUTXO(address)

    for _, out := range UTXOs {
        balance += out.Value
    }

    fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
    chain := blockchain.ContinueBlockChain(from)
    defer chain.Database.Close()

    tx := blockchain.NewTransaction(from, to, amount, chain)
    chain.AddBlock([]*blockchain.Transaction{tx})
    fmt.Println("Success!")
}

func (cli *CommandLine) createWallet() {
    wallets, _ := wallet.CreateWallets()
    addresses := wallets.AddWallet()
    wallets.SaveFile()

    fmt.Printf("Your wallet address is: %s\n", address)
}


func (cli *CommandLine) listAddresses() {
    wallets, _ := wallet.CreateWallets()
    addresses := wallets.GetAllAddresses()

    for _, address L= range addresses {
        fmt.Println(address)
    }
}

func (cli *CommandLine) Run() {
    cli.validateArgs()

    getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
    createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
    sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
    printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
    createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
    getWalletsCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)

    getBalanceAddress := getBalanceCmd.String("address", "", "The address to check")
    createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send rewards to")
    sendFrom := sendCmd.String("from", "", "source wallet")
    sendTo := sendCmd.String("to", "", "destination wallet")
    sendAmount := sendCmd.Int("amount", 0, "Amount to send")

    switch os.Args[1] {
    case "getbalance":
        err := getBalanceCmd.Parse(os.Args[2:])
        blockchain.Handle(err)
    case "createblockchain":
        err := createBlockchainCmd.Parse(os.Args[2:])
        blockchain.Handle(err)
    case "send":
        err := sendCmd.Parse(os.Args[2:])
        blockchain.Handle(err)
    case "printchain":
        err := printChainCmd.Parse(os.Args[2:])
        blockchain.Handle(err)
    case "createwallet":
        err := createWalletCmd.Parse(os.Args[2:])
        blockchain.Handle(err)
    case "getallwallets":
        err := getWalletsCmd.Parse(os.Args[2:])
        blockchain.Handle(err)
    default:
        cli.printUsage()
        runtime.Goexit()
    }

    if getBalanceCmd.Parsed() {
        if *getBalanceAddress == "" {
            getBalanceCmd.Usage()
            runtime.Goexit()
        }
        cli.getBalance(*getBalanceAddress)
    }
    
    if createBlockchainCmd.Parsed() {
        if *createBlockchainAddress == "" {
            createBlockchainCmd.Usage()
            runtime.Goexit()
        }
        cli.createBlockChain(*createBlockchainAddress)
    }
    if sendCmd.Parsed() {
        if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
            sendCmd.Usage()
            runtime.Goexit()
        }
        cli.send(*sendFrom, *sendTo, *sendAmount)
    }
    
    if printChainCmd.Parsed() {
        cli.printChain()
    }
    
    if createWalletCmd.Parsed() {
        cli.createWallet()
    }

    if listAddressesCmd.Parsed() {
        cli.listAddresses()
    }

}

