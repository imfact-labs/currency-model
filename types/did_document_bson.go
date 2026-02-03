package types

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (d DIDDocument) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"_hint":              d.Hint().String(),
		"@context":           d.context_,
		"id":                 d.id,
		"authentication":     d.authentication,
		"verificationMethod": d.verificationMethod,
		"service":            d.service,
	})
}

type DIDDocumentBSONUnmarshaler struct {
	Hint      string                    `bson:"_hint"`
	Context_  []string                  `bson:"@context"`
	ID        string                    `bson:"id"`
	Auth      []VerificationMethodOrRef `bson:"authentication"`
	VRFMethod bson.Raw                  `bson:"verificationMethod"`
	Service   []Service                 `bson:"service"`
}

func (d *DIDDocument) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of DIDDocument")

	var u DIDDocumentBSONUnmarshaler
	if err := bsonenc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	d.BaseHinter = hint.NewBaseHinter(ht)

	authSlice := make([]VerificationRelationshipEntry, 0)
	if len(u.Auth) > 0 {
		for _, v := range u.Auth {
			authSlice = append(authSlice, &v)
		}
	}

	d.authentication = authSlice

	hr, err := enc.DecodeSlice(u.VRFMethod)
	if err != nil {
		return err
	}

	vrfs := make([]IVerificationMethod, len(hr))
	if len(hr) > 0 {
		for i, hinter := range hr {
			if v, ok := hinter.(IVerificationMethod); !ok {
				return e.Wrap(errors.Errorf("expected DIDVerificationMethod, not %T", hinter))
			} else {
				if err := v.IsValid(nil); err != nil {
					return e.Wrap(err)
				} else {
					vrfs[i] = v
				}
			}

		}
	}

	d.verificationMethod = vrfs
	if u.Service != nil {
		d.service = u.Service
	} else {
		d.service = []Service{}
	}

	return d.unpack(u.Context_, u.ID)
}

func (d Service) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"id":                d.id.String(),
		"type":              d.serviceType,
		"service_end_point": d.serviceEndPoint,
	})
}

type ServiceBSONUnmarshaler struct {
	ID              string `bson:"id"`
	Type            string `bson:"type"`
	ServiceEndPoint string `bson:"service_end_point"`
}

func (d *Service) UnmarshalBSON(b []byte) error {
	e := util.StringError("decode bson of Service")

	var u ServiceBSONUnmarshaler
	err := bsonenc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	return d.unpack(u.ID, u.Type, u.ServiceEndPoint)
}
