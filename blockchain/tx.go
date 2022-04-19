package blockchain

import (
    "bytes"
    "github.com/viscory/reciprocus/wallet"
    "encoding/gob"
)

type Transaction struct {
    ID []byte
    Inputs []TxInput
    Outputs []TxOutput
}

type TxOutput struct {
    Value int
    PubKeyHash []byte
}

type TxOutputs struct {
    Outputs []TxOutput
}

type TxInput struct {
    ID []byte
    Out int
    Signature []byte
    PubKey []byte 
}

func NewTxOutput(value int, address string) *TxOutput {
    txo := &TxOutput{value, nil}
    txo.Lock([]byte(address))

    return txo
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
    lockingHash := wallet.PublicKeyHash(in.PubKey)
    return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
    pubKeyHash := wallet.Base58Decode(address)
    pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]
    out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
    return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func (outs TxOutputs) Serialize() []byte {
    var buffer bytes.Buffer
    encode := gob.NewEncoder(&buffer)
    err := encode.Encode(outs)
    Handle(err)
    return buffer.Bytes()
}

func DeserializeOututs(data []byte) TxOutputs {
    var outputs TxOutputs
    decode := gob.NewDecoder(bytes.NewReader(data))
    err := decode.Decode(&outputs)
    Handle(err)
    return outputs
}
