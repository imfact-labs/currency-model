package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/utils/bsonenc"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (ca Address) MarshalBSONValue() (byte, []byte, error) {
	typ, data, err := bson.MarshalValue(ca.String())
	if err != nil {
		return 0, nil, err
	}

	return byte(typ), data, nil
}

func (ca *Address) DecodeBSON(b []byte, _ *bsonenc.Encoder) error {
	*ca = NewAddress(string(b))

	return nil
}
