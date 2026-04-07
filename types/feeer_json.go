package types

import (
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
)

func (fa NilFeeer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(hint.BaseHinter{
		HT: fa.Hint(),
	})
}

func (fa *NilFeeer) UnmarsahlJSON(b []byte) error {
	e := util.StringError("unmarshal json of NilFeeer")

	var ht hint.BaseHinter
	if err := util.UnmarshalJSON(b, &ht); err != nil {
		return e.Wrap(err)
	}

	fa.BaseHinter = ht

	return nil
}

type FixedFeeerJSONMarshaler struct {
	hint.BaseHinter
	Receiver base.Address `json:"receiver"`
	Amount   string       `json:"amount"`
}

func (fa FixedFeeer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FixedFeeerJSONMarshaler{
		BaseHinter: fa.BaseHinter,
		Receiver:   fa.receiver,
		Amount:     fa.amount.String(),
	})
}

type FixedFeeerJSONUnmarshaler struct {
	Hint     hint.Hint `json:"_hint"`
	Receiver string    `json:"receiver"`
	Amount   string    `json:"amount"`
}

func (fa *FixedFeeer) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("Decode json of FixedFeeer")

	var ufa FixedFeeerJSONUnmarshaler
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return e.Wrap(err)
	}

	return fa.unpack(enc, ufa.Hint, ufa.Receiver, ufa.Amount)
}

type FixedItemDataSizeExecutionFeeerJSONMarshaler struct {
	hint.BaseHinter
	Receiver           base.Address `json:"receiver"`
	Amount             string       `json:"amount"`
	ItemFeeAmount      string       `json:"item_fee_amount"`
	DataSizeFeeAmount  string       `json:"data_size_fee_amount"`
	DataSizeUnit       int64        `json:"data_size_unit"`
	ExecutionFeeAmount string       `json:"execution_fee_amount"`
}

func (fa FixedItemDataSizeExecutionFeeer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FixedItemDataSizeExecutionFeeerJSONMarshaler{
		BaseHinter:         fa.BaseHinter,
		Receiver:           fa.receiver,
		Amount:             fa.amount.String(),
		ItemFeeAmount:      fa.itemFeeAmount.String(),
		DataSizeFeeAmount:  fa.dataSizeFeeAmount.String(),
		DataSizeUnit:       fa.dataSizeUnit,
		ExecutionFeeAmount: fa.executionFeeAmount.String(),
	})
}

type FixedItemDataSizeExecutionFeeerJSONUnmarshaler struct {
	Hint               hint.Hint `json:"_hint"`
	Receiver           string    `json:"receiver"`
	Amount             string    `json:"amount"`
	ItemFeeAmount      string    `json:"item_fee_amount"`
	DataSizeFeeAmount  string    `json:"data_size_fee_amount"`
	DataSizeUnit       int64     `json:"data_size_unit"`
	ExecutionFeeAmount string    `json:"execution_fee_amount"`
}

func (fa *FixedItemDataSizeExecutionFeeer) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("Decode json of FixedItemDataSizeFeeer")

	var ufa FixedItemDataSizeExecutionFeeerJSONUnmarshaler
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return e.Wrap(err)
	}

	return fa.unpack(
		enc, ufa.Hint, ufa.Receiver, ufa.Amount, ufa.ItemFeeAmount,
		ufa.DataSizeFeeAmount, ufa.DataSizeUnit, ufa.ExecutionFeeAmount,
	)
}
