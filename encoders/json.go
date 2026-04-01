package encoders

import (
	"encoding/json"
)

type JsonEncoder struct{}

func NewJsonEncoder() Encoder {
	return &JsonEncoder{}
}

func (e JsonEncoder) Encode(v any) ([]byte, error)    { return json.Marshal(v) }
func (e JsonEncoder) Decode(data []byte, v any) error { return json.Unmarshal(data, v) }
func (e JsonEncoder) Name() string                    { return "json" }
