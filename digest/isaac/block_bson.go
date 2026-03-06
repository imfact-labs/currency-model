package isaac

import (
	"time"

	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (m Manifest) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":           m.Hint().String(),
			"proposed_at":     m.proposedAt,
			"states_tree":     m.statesTree,
			"hash":            m.h,
			"previous":        m.previous,
			"proposal":        m.proposal,
			"operations_tree": m.operationsTree,
			"suffrage":        m.suffrage,
			"height":          m.height,
		},
	)
}

type ManifestBSONUnmarshaler struct {
	Hint           string          `bson:"_hint"`
	ProposedAt     time.Time       `bson:"proposed_at"`
	StatesTree     valuehash.Bytes `bson:"states_tree"`
	Hash           valuehash.Bytes `bson:"hash"`
	Previous       valuehash.Bytes `bson:"previous"`
	Proposal       valuehash.Bytes `bson:"proposal"`
	OperationsTree valuehash.Bytes `bson:"operations_tree"`
	Suffrage       valuehash.Bytes `bson:"suffrage"`
	Height         base.Height     `bson:"height"`
}

func (m *Manifest) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of Manifest")

	var u ManifestBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	m.BaseHinter = hint.NewBaseHinter(ht)
	m.h = u.Hash
	m.height = u.Height
	m.previous = u.Previous
	m.proposal = u.Proposal
	m.operationsTree = u.OperationsTree
	m.statesTree = u.StatesTree
	m.suffrage = u.Suffrage
	m.proposedAt = u.ProposedAt

	return nil
}
