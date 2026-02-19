package currency

import (
	"encoding/json"
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

type TransferItemJSONPacker struct {
	hint.BaseHinter
	Receiver base.Address   `json:"receiver"`
	Amounts  []types.Amount `json:"amounts"`
}

func (it BaseTransferItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(TransferItemJSONPacker{
		BaseHinter: it.BaseHinter,
		Receiver:   it.receiver,
		Amounts:    it.amounts,
	})
}

type BaseTransferItemJSONUnpacker struct {
	Hint     hint.Hint       `json:"_hint"`
	Receiver string          `json:"receiver"`
	Amounts  json.RawMessage `json:"amounts"`
}

func (it *BaseTransferItem) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uit BaseTransferItemJSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	if err := it.unpack(enc, uit.Hint, uit.Receiver, uit.Amounts); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	return nil
}
