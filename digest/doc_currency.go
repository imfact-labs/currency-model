package digest

import (
	mongodbst "github.com/imfact-labs/currency-model/digest/mongodb"
	"github.com/imfact-labs/currency-model/state/currency"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/pkg/errors"
)

type CurrencyDoc struct {
	mongodbst.BaseDoc
	st base.State
	cd types.CurrencyDesign
}

func NewCurrencyDoc(st base.State, enc encoder.Encoder) (CurrencyDoc, error) {
	cd, err := currency.GetDesignFromState(st)
	if err != nil {
		return CurrencyDoc{}, errors.Wrap(err, "CurrencyDoc needs CurrencyDesign state")
	}

	b, err := mongodbst.NewBaseDoc(nil, st, enc)
	if err != nil {
		return CurrencyDoc{}, err
	}

	return CurrencyDoc{
		BaseDoc: b,
		st:      st,
		cd:      cd,
	}, nil
}

func (doc CurrencyDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["currency"] = doc.cd.Currency().String()
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}
