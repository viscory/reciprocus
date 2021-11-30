package blockchain

import (
    "bytes"
    "crypto/sha256"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "encoding/gob"
    "encoding/hex"
    "fmt"
    "log"
    "math/big"
    "strings"
    "github.com/viscory/reciprocus/wallet"
)

func (tx Transaction) Serialize() []byte {
    var encoded bytes.Buffer

    enc := gob.NewEncoder(&encoded)
    err := enc.Encode(tx)
    if err != nil {
        log.Panic(err)
    }

    return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
    var hash [32]byte

    txCopy := *tx
    txCopy.ID = []byte{}

    hash = sha256.Sum256(txCopy.Serialize())

    return hash[:]
}

func (tx* Transaction) SetID() {
    var encoded bytes.Buffer
    var hash [32]byte
    
    encode := gob.NewEncoder(&encoded)
    err := encode.Encode(tx)
    Handle(err)

    hash = sha256.Sum256(encoded.Bytes())
    tx.ID = hash[:]
}

func CoinbaseTx(to, data string) *Transaction {
    if data == "" {
        data = fmt.Sprintf("Coins to %s", to)
    }


    txin := TxInput{[]byte{}, -1, []byte(data)}
    txout := NewTXOutput(100, to)

    tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}
    tx.SetID()

    return &tx
}

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
    var inputs []TxInput
    var outputs []TxOutput

    wallets, err := wallet.CreateWallets()
    Handle(err)
    w := wallets.GetWallet(from)
    pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

    acc, validOutputs := chain.FindSpendableOutputs(from, amount)

    if acc < amount {
        log.Panic("Not enough funds")
    }

    for txid, outs := range validOutputs {
        txID, err := hex.DecodeString(txid)
        Handle(err)

        for _, out := range outs {
            input := TxInput{txID, out, w.PublicKey}
            inputs = append(inputs, input)
        }
    }

    outputs = append(outputs, *NewTXOutput(amount, to))

    if acc > amount {
        outputs = append(outputs, *NewTXOutput(acc-amount, from))
    }

    tx := Transaction{nil, inputs, outputs}
    tx.ID = tx.Hash()
    chain.SignTransaction(&tx, w.PrivateKey)

    return &tx
}

func (tx *Transaction) IsCoinBase() bool {
    return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
    if tx.IsCoinbase() {
        return
    }

    for _, in := range tx.Inputs {
        if prevTXs[hex.EncodeToString(in.id)].ID== nil {
            log.Panic("Previous Transaction Does Not Exist")
        }
    }

    txCopy := tx.TrimmedCopy()

    for inId, in := range txCoy.Inputs {
        prevTX := prevTXs[hex.EncodeToString(in.ID)]
        txCopy.Inputs[inId].Signature = nil
        txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.out].PubKeyHash
        txCopy.ID = txCopy.Hash()
        txCopy.Inputs[inId].PubKey = nil

        r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
        Handle(err)
        signature := append(r.Bytes(), s.Bytes()...)

        tx.Inputs[inId].Signature = signature
    }
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
    if tx.IsCoinbase() {
        return true
    }
    
    for _, in := range tx.Inputs {
        if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
            log.Panic("Previous transaction doesn't exist")
        }
    }

    txCopy := tx.TrimmedCopy()
    curve := elliptic.P256()

    for inId, in := range tx.Inputs {
        prevTx := prevTXs[hex.EncodeToString(in.ID)]
        txCopy.Inputs[inId].Signature = nil
        txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash
        txCopy.ID = txCopy.Hash()
        txCopy.Inputs[inId].PubKey = nil

        r := big.Int{}
        s := big.Int{}
        sigLen := len(in.Signature)
        r.setBytes(in.Signature[:(sigLen/2)])
        s.setBytes(in.Signature[(sigLen/2):])

        x := big.Int{}
        y := big.Int{}
        keyLen - len(in.PubKey)
        x.setBytes(in.Signature[:(keyLen/2)])
        y.setBytes(in.Signature[(keyLen/2):])

        rawPubKey := ecdsa.PublicKey{curse, &x, &y}
        if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
            return false
        }
    }
    return true
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
    iter := bc.Iterator()
    
    for {
        block := iter.Next()

        for _, tx := range block.Transactions {
            if bytes.Compare(tx.ID, ID) == 0 {
                return *tx, nil
            }           
        }

        if len(block.PrevHash) == 0 {
            break
        }
    }
    return Transaction{} , errors.New("Transaction does not exist")
}

func (tx *Transaction) TrimmedCopy() Transaction {
    var inputs []TxInput
    var outputs []TxOutput

    for _, in := range tx.Inputs {
        inputs = append(inputs, TxInput{in.Id, in.Out, nil, nil})
    }
    
    for _, out := range tx.Outputs {
        outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
    }

    txCopy := Transaction{tx.ID, inputs, outputs}

    return txCopy
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
