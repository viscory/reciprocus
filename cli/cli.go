package cli

import (
    "github.com/viscory/reciprocus/blockchain"
    "github.com/viscory/reciprocus/wallet"
    "github.com/viscory/reciprocus/network"
    
    "fmt"
    "flag"
    "os"
    "runtime"
    "log"
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
    fmt.Println(" getalwallets - lists all wallets inside wallet file")
    fmt.Println(" reindexutxo - reindexes utxo set")
    fmt.Println(" startminer -miner ADDRESS - start a node with id specified as $NODE_ID")
}

func (cli *CommandLine) validateArgs() {
    if len(os.Args) < 2 {
        cli.printUsage()
        runtime.Goexit()
    }
}

func (cli *CommandLine) printChain(nodeId string) {
    chain := blockchain.ContinueBlockChain(nodeId)
    defer chain.Database.Close()
    iter := chain.Iterator()

    for {
        block := iter.Next()

        fmt.Printf("Previous Hash: %x\n", block.PrevHash)
        fmt.Printf("Hash: %x\n", block.Hash)
        pow:= blockchain.NewProof(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
        for _, tx := range block.Transactions {
            fmt.Println(tx)
        }
        fmt.Println()

        if len(block.PrevHash) == 0 {
            break
        }
    }
}

func (cli *CommandLine) createBlockChain(address, nodeId string) {
    if !wallet.ValidateAddress(address) {
        log.Panic("Address is not Valid")
    }
    chain := blockchain.InitBlockChain(address, nodeId)
    chain.Database.Close()

    UTXOSet := blockchain.UTXOSet{chain}
    UTXOSet.Reindex()
    fmt.Println("Created Blockchain")
}

func (cli *CommandLine) getBalance(address, nodeId string) {
    if !wallet.ValidateAddress(address) {
        log.Panic("Address is not Valid")
    }
    chain := blockchain.ContinueBlockChain(nodeId)
    UTXOSet := blockchain.UTXOSet{chain}
    defer chain.Database.Close()

    balance := 0
    pubKeyHash := wallet.Base58Decode([]byte(address))
    pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]
    UTXOs := UTXOSet.FindUTXO(pubKeyHash)

    for _, out := range UTXOs {
        balance += out.Value
    }

    fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int, nodeId string, mineNow bool ) {
    if !wallet.ValidateAddress(to) {
        log.Panic("Address is not valid!")
    }
    if !wallet.ValidateAddress(from) {
        log.Panic("Address is not valid!")
    }

    chain := blockchain.ContinueBlockChain(nodeId)
    UTXOSet := blockchain.UTXOSet{chain}
    defer chain.Database.Close() 

    wallets, err := wallet.CreateWallets(nodeId)
    if err != nil {
        log.Panic(err)
    }
    wallet := wallets.GetWallet(from)

    tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)
    if mineNow {
        cbTx := blockchain.CoinbaseTx(from, "")
        txs := []*blockchain.Transaction{cbTx, tx}
        block := chain.MineBlock(txs)
        UTXOSet.Update(block)
    } else {
        network.SendTx(network.KnownNodes[0], tx)
        fmt.Println("send tx")
    }

    fmt.Println("Success!")
}
 
func (cli *CommandLine) createWallet(nodeId string) {
    wallets, _ := wallet.CreateWallets(nodeId)
    address := wallets.AddWallet()
    wallets.SaveFile(nodeId)

    fmt.Printf("Your wallet address is: %s\n", address)
}


func (cli *CommandLine) listAddresses(nodeId string) {
    wallets, _ := wallet.CreateWallets(nodeId )
    addresses := wallets.GetAllAddresses()

    for _, address := range addresses {
        fmt.Println(address)
    }
}

func (cli *CommandLine) reindexUTXO(nodeId string) {
    chain := blockchain.ContinueBlockChain(nodeId)
    defer chain.Database.Close()
    UTXOSet := blockchain.UTXOSet{chain}
    UTXOSet.Reindex()

    count := UTXOSet.CountTransactions()
    fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CommandLine) StartNode(nodeId, minerAddress string) {
    fmt.Printf("Starting Node %s\n", nodeId)

    if len(minerAddress) > 0 {
        if wallet.ValidateAddress(minerAddress) {
            fmt.Println("Mining on, receiving reward in: ", minerAddress)
        } else {
            log.Panic("Invalid miner address")
        }
    }
    network.StartServer(nodeId, minerAddress)
}

func (cli *CommandLine) Run() {
    cli.validateArgs()

    nodeId := os.Getenv("NODE_ID")
    if nodeId == "" {
        fmt.Printf("$NODE_ID not set")
        runtime.Goexit()
    }

    getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
    createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
    sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
    printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
    createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
    getWalletsCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
    reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
    startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

    getBalanceAddress := getBalanceCmd.String("address", "", "The address to check")
    createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send rewards to")
    sendFrom := sendCmd.String("from", "", "source wallet")
    sendTo := sendCmd.String("to", "", "destination wallet")
    sendAmount := sendCmd.Int("amount", 0, "Amount to send")
    sendMine := sendCmd.Bool("mine", false, "mine immediately on the same node")
    startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to address")

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
    case "reindexutxo":
        err := reindexUTXOCmd.Parse(os.Args[2:])
        blockchain.Handle(err)
    case "startnode":
        err := startNodeCmd.Parse(os.Args[2:])
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
        cli.getBalance(*getBalanceAddress, nodeId)
    }
    
    if createBlockchainCmd.Parsed() {
        if *createBlockchainAddress == "" {
            createBlockchainCmd.Usage()
            runtime.Goexit()
        }
        cli.createBlockChain(*createBlockchainAddress, nodeId)
    }
    if sendCmd.Parsed() {
        if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
            sendCmd.Usage()
            runtime.Goexit()
        }
        cli.send(*sendFrom, *sendTo, *sendAmount, nodeId, *sendMine)
    }
    
    if printChainCmd.Parsed() {
        cli.printChain(nodeId)
    }
    
    if createWalletCmd.Parsed() {
        cli.createWallet(nodeId)
    }

    if getWalletsCmd.Parsed() {
        cli.listAddresses(nodeId)
    }
    if reindexUTXOCmd.Parsed() {
        cli.reindexUTXO(nodeId)
    }
    if startNodeCmd.Parsed() {
        nodeId := os.Getenv("NODE_ID")
        if nodeId == "" {
            startNodeCmd.Usage()
            runtime.Goexit()
        }
        cli.StartNode(nodeId, *startNodeMiner)
    }

}

