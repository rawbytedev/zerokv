package encoders

type Encoder interface {
	Encode(val any) ([]byte, error)
	Decode(data []byte, val any) error
	Name() string
}
