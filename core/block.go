package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/LeiZhou-97/blockchain/crypto"
	"github.com/LeiZhou-97/blockchain/types"
)

type Header struct {
	Version       uint32
	DataHash      types.Hash
	PrevBlockHash types.Hash
	Timestamp     int64
	Height        uint32
	Nonce         uint64
}

//	func (h *Header) EncodeBinary(w io.Writer) error {
//	    if err := binary.Write(w, binary.LittleEndian, &h.Version); err != nil {
//	        return err
//	    }
//	    if err := binary.Write(w, binary.LittleEndian, &h.PrevBlock); err != nil {
//	        return err
//	    }
//	    if err := binary.Write(w, binary.LittleEndian, &h.Timestamp); err != nil {
//	        return err
//	    }
//	    if err := binary.Write(w, binary.LittleEndian, &h.Height); err != nil {
//	        return err
//	    }
//	    if err := binary.Write(w, binary.LittleEndian, &h.Nonce); err != nil {
//	        return err
//	    }
//	    if err := binary.Write(w, binary.LittleEndian, &h.Version); err != nil {
//	        return err
//	    }
//	    return binary.Write(w, binary.LittleEndian, &h.Nonce)
//	}
//
//	func (h *Header) DecodeBinary(r io.Reader) error {
//	    if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
//	        return err
//	    }
//	    if err := binary.Read(r, binary.LittleEndian, &h.PrevBlock); err != nil {
//	        return err
//	    }
//	    if err := binary.Read(r, binary.LittleEndian, &h.Timestamp); err != nil {
//	        return err
//	    }
//	    if err := binary.Read(r, binary.LittleEndian, &h.Height); err != nil {
//	        return err
//	    }
//	    if err := binary.Read(r, binary.LittleEndian, &h.Nonce); err != nil {
//	        return err
//	    }
//	    if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
//	        return err
//	    }
//	    return binary.Read(r, binary.LittleEndian, &h.Nonce)
//
// }
type Block struct {
	*Header
	Transactions []Transaction
	Validator    crypto.PublicKey
	Signature    *crypto.Signature
	// cached version of the header hash
	hash types.Hash
}

func NewBlock(h *Header, txx []Transaction) *Block {
	return &Block{
		Header:       h,
		Transactions: txx,
	}
}

func (b *Block) Sign(privKey crypto.PrivateKey) error {
	sig, err := privKey.Sign(b.HeaderData())	
	if err != nil {
		panic(err)
	}

	b.Validator = privKey.PublicKey()
	b.Signature = sig
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("block has no sign")
	}

	if !b.Signature.Verify(b.Validator, b.HeaderData()) {
		return fmt.Errorf("block has invalid sign")
	}

	return nil 	
}

func (b *Block) Decode(r io.Reader, dec Decoder[*Block]) error {
	return dec.Decode(r, b)
}

func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}

func (b *Block) Hash(hasher Hasher[*Block]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b)
	}
	return b.hash
}


func (b *Block) HeaderData() []byte {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	enc.Encode(b.Header)
	return buf.Bytes()
}

// func (b *Block) Hash() types.Hash {
//     buf := &bytes.Buffer{}
//     b.Header.EncodeBinary(buf)
//
//     if b.hash.IsZero() {
//         b.hash = types.Hash(sha256.Sum256(buf.Bytes()))
//     }
//
//     return b.hash
// }
//
// func (b *Block) EncodeBinary(w io.Writer) error {
//     if err := b.Header.EncodeBinary(w); err != nil {
//         return err
//     }
//
//     for _,tx := range b.Transactions {
//         if err := tx.EncodeBinary(w); err != nil {
//             return err
//         }
//     }
//     return nil
// }
//
// func (b *Block) DecodeBinary(r io.Reader) error {
//     if err := b.Header.DecodeBinary(r); err != nil {
//         return err
//     }
//
//     for _,tx := range b.Transactions {
//         if err := tx.DecodeBinary(r); err != nil {
//             return err
//         }
//     }
//     return nil
// }

