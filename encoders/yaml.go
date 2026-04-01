package encoders

import (
	"gopkg.in/yaml.v3"
)

type YamlEncoder struct{}

func NewYamlEncoder() Encoder {
	return &YamlEncoder{}
}

func (e YamlEncoder) Encode(v any) ([]byte, error)    { return yaml.Marshal(v) }
func (e YamlEncoder) Decode(data []byte, v any) error { return yaml.Unmarshal(data, v) }
func (e YamlEncoder) Name() string                    { return "yaml" }
