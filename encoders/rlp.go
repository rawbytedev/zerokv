package encoders

import (
	"github.com/ethereum/go-ethereum/rlp"
)

type RLPEncoder struct{}

func NewRLPEncoder() Encoder {
	return &RLPEncoder{}
}

func (e RLPEncoder) Encode(v any) ([]byte, error)    { return rlp.EncodeToBytes(v) }
func (e RLPEncoder) Decode(data []byte, v any) error { return rlp.DecodeBytes(data, v) }
func (e RLPEncoder) Name() string                    { return "rlp" }
