package steps

import (
	"context"
	
	"github.com/imfact-labs/currency-model/app/runtime/contracts"
	"github.com/imfact-labs/currency-model/app/runtime/spec"
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/encoder/json"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

var BEncoderContextKey = util.ContextKey("bson-encoder")

func PEncoder(pctx context.Context) (context.Context, error) {
	e := util.StringError("prepare encoders")

	jenc := jsonenc.NewEncoder()
	encs := encoder.NewEncoders(jenc, jenc)
	benc := bsonenc.NewEncoder()

	if err := encs.AddEncoder(benc); err != nil {
		return pctx, e.Wrap(err)
	}

	return util.ContextWithValues(pctx, map[util.ContextKey]interface{}{
		launch.EncodersContextKey: encs,
		BEncoderContextKey:        benc,
	}), nil
}

func PAddHinters(pctx context.Context) (context.Context, error) {
	e := util.StringError("add hinters")

	var encs *encoder.Encoders
	var f contracts.ProposalOperationFactHintFunc = IsSupportedProposalOperationFactHintFunc

	if err := util.LoadFromContextOK(pctx, launch.EncodersContextKey, &encs); err != nil {
		return pctx, e.Wrap(err)
	}
	pctx = context.WithValue(pctx, contracts.ProposalOperationFactHintContextKey, f)

	if err := LoadHinters(encs); err != nil {
		return pctx, e.Wrap(err)
	}

	return pctx, nil
}

func IsSupportedProposalOperationFactHintFunc() func(hint.Hint) bool {
	return func(ht hint.Hint) bool {
		for i := range spec.SupportedProposalOperationFactHinters {
			s := spec.SupportedProposalOperationFactHinters[i].Hint
			if ht.Type() != s.Type() {
				continue
			}

			return ht.IsCompatible(s)
		}

		return false
	}
}

func LoadHinters(encs *encoder.Encoders) error {
	for i := range spec.Hinters {
		if err := encs.AddDetail(spec.Hinters[i]); err != nil {
			return errors.Wrap(err, "add hinter to encoder")
		}
	}

	for i := range spec.SupportedProposalOperationFactHinters {
		if err := encs.AddDetail(spec.SupportedProposalOperationFactHinters[i]); err != nil {
			return errors.Wrap(err, "add supported proposal operation fact hinter to encoder")
		}
	}

	return nil
}
