package dataTypes

import "io"

type Array struct {
	TypeCode int

	data []DataType
}

func (a *Array) Type() int {
	return a.TypeCode
}

func (a *Array) Serialize(encoder Encoder) error {
	if err := encoder.writeByte('*'); err != nil {
		return err
	}
	for i := 0; i < len(a.data); i++ {
		a.data[i].Serialize(encoder)
	}
	return nil
}

func (a *Array) Deserialization(decoder io.Reader) (DataType, error) {

}
