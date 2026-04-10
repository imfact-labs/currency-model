package types

import (
	"github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (r BaseFeeReceipt) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":       r.Hint().String(),
			"currency_id": r.CurrencyID,
			"amount":      r.Amount,
		},
	)
}

type BaseFeeReceiptBSONUnmarshaler struct {
	Hint       string     `bson:"_hint"`
	CurrencyID CurrencyID `bson:"currency_id"`
	Amount     string     `bson:"amount"`
}

func (r *BaseFeeReceipt) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u BaseFeeReceiptBSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	hts := u.Hint
	if hts == "" {
		hts = BaseFeeReceiptHint.String()
	}

	ht, err := hint.ParseHint(hts)
	if err != nil {
		return err
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.CurrencyID = u.CurrencyID
	r.Amount = u.Amount

	return nil
}

func (r FixedFeeReceipt) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":       r.Hint().String(),
			"currency_id": r.CurrencyID,
			"amount":      r.Amount,
			"base_amount": r.BaseAmount,
		},
	)
}

type FixedFeeReceiptBSONUnmarshaler struct {
	Hint       string     `bson:"_hint"`
	CurrencyID CurrencyID `bson:"currency_id"`
	Amount     string     `bson:"amount"`
	BaseAmount string     `bson:"base_amount"`
}

func (r *FixedFeeReceipt) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u FixedFeeReceiptBSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	hts := u.Hint
	if hts == "" {
		hts = FixedFeeReceiptHint.String()
	}

	ht, err := hint.ParseHint(hts)
	if err != nil {
		return err
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.CurrencyID = u.CurrencyID
	r.Amount = u.Amount
	r.BaseAmount = u.BaseAmount

	return nil
}

func (r FixedItemDataSizeExecutionFeeReceipt) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":                 r.Hint().String(),
			"currency_id":           r.CurrencyID,
			"total_amount":          r.TotalAmount,
			"base_amount":           r.BaseAmount,
			"item_count":            r.ItemCount,
			"item_fee_amount":       r.ItemFeeAmount,
			"item_fee":              r.ItemFee,
			"data_size":             r.DataSize,
			"data_size_unit":        r.DataSizeUnit,
			"data_size_fee_amount":  r.DataSizeFeeAmount,
			"data_size_fee":         r.DataSizeFee,
			"execution_fee_amount":  r.ExecutionFeeAmount,
			"execution_fee":         r.ExecutionFee,
		},
	)
}

type FixedItemDataSizeExecutionFeeReceiptBSONUnmarshaler struct {
	Hint               string     `bson:"_hint"`
	CurrencyID         CurrencyID `bson:"currency_id"`
	TotalAmount        string     `bson:"total_amount"`
	BaseAmount         string     `bson:"base_amount"`
	ItemCount          int        `bson:"item_count"`
	ItemFeeAmount      string     `bson:"item_fee_amount"`
	ItemFee            string     `bson:"item_fee"`
	DataSize           int        `bson:"data_size"`
	DataSizeUnit       int64      `bson:"data_size_unit"`
	DataSizeFeeAmount  string     `bson:"data_size_fee_amount"`
	DataSizeFee        string     `bson:"data_size_fee"`
	ExecutionFeeAmount string     `bson:"execution_fee_amount"`
	ExecutionFee       string     `bson:"execution_fee"`
}

func (r *FixedItemDataSizeExecutionFeeReceipt) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u FixedItemDataSizeExecutionFeeReceiptBSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	hts := u.Hint
	if hts == "" {
		hts = FixedItemDataSizeExecutionFeeReceiptHint.String()
	}

	ht, err := hint.ParseHint(hts)
	if err != nil {
		return err
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.CurrencyID = u.CurrencyID
	r.TotalAmount = u.TotalAmount
	r.BaseAmount = u.BaseAmount
	r.ItemCount = u.ItemCount
	r.ItemFeeAmount = u.ItemFeeAmount
	r.ItemFee = u.ItemFee
	r.DataSize = u.DataSize
	r.DataSizeUnit = u.DataSizeUnit
	r.DataSizeFeeAmount = u.DataSizeFeeAmount
	r.DataSizeFee = u.DataSizeFee
	r.ExecutionFeeAmount = u.ExecutionFeeAmount
	r.ExecutionFee = u.ExecutionFee

	return nil
}

func (r CurrencyOperationReceipt) MarshalBSON() ([]byte, error) {
	m := bson.M{
		"_hint": r.Hint().String(),
	}

	if r.Fee != nil {
		m["fee"] = r.Fee
	}

	if r.GasUsed != nil {
		m["gas_used"] = r.GasUsed
	}

	return bsonenc.Marshal(m)
}

type CurrencyOperationReceiptBSONUnmarshaler struct {
	Hint    string   `bson:"_hint"`
	Fee     bson.Raw `bson:"fee"`
	GasUsed *uint64  `bson:"gas_used"`
}

func (r *CurrencyOperationReceipt) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u CurrencyOperationReceiptBSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	hts := u.Hint
	if hts == "" {
		hts = CurrencyOperationReceiptHint.String()
	}

	ht, err := hint.ParseHint(hts)
	if err != nil {
		return err
	}

	r.BaseHinter = hint.NewBaseHinter(ht)
	r.GasUsed = u.GasUsed

	if len(u.Fee) < 1 {
		r.Fee = nil

		return nil
	}

	var fee FeeReceipt
	if err := encoder.Decode(enc, u.Fee, &fee); err != nil {
		return err
	}

	r.Fee = fee

	return nil
}
