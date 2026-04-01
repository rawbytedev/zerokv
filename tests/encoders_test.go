package tests

import (
	"fmt"
	"testing"

	"github.com/rawbytedev/zerokv/helpers"
	"github.com/stretchr/testify/require"
)

type TestingStruct struct {
	Name string
	Data []byte
}

func GenerateRandom() TestingStruct {
	return TestingStruct{Name: "ZeroKv", Data: []byte{0x01, 0x69, 0x52, 0x75}}
}

func TestEncoders(t *testing.T) {
	encoder := []string{
		"json", "yaml", "rlp",
	}
	test_list := []test{
		{
			name: "Encode",
			fn:   testEncoding,
		},
	}
	for i := range encoder {
		for tt := range test_list {
			testname := fmt.Sprintf("%s_%s", test_list[tt].name, encoder[i])
			t.Run(testname, func(t *testing.T) {
				test_list[tt].fn(t, encoder[i])
			})
		}
	}
}

func testEncoding(t *testing.T, name string) {
	field := GenerateRandom()
	field2 := &TestingStruct{}
	enc := helpers.SetupEncoders(t, name)
	data, err := enc.Encode(field)
	require.NoError(t, err)
	err = enc.Decode(data, field2)
	require.NoError(t, err)
	require.EqualExportedValues(t, field, *field2)
}
