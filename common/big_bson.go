package common

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/x/bsonx/bsoncore"
)

func (a Big) MarshalBSONValue() (byte, []byte, error) {
	typ, data, err := bson.MarshalValue(a.String())
	if err != nil {
		return 0, nil, err
	}

	return byte(typ), data, nil
}

func (a *Big) UnmarshalBSONValue(t byte, b []byte) error {
	if bson.Type(t) != bson.TypeString {
		return errors.Errorf("Invalid marshaled type for Big, %v", bson.Type(t))
	}

	s, _, ok := bsoncore.ReadString(b)
	if !ok {
		return errors.Errorf("Can not read string")
	}

	ua, err := NewBigFromString(s)
	if err != nil {
		return err
	}
	*a = ua

	return nil
}
