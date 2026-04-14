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
			"currency_id": r.currencyID,
			"total_fee":   r.totalFee,
		},
	)
}

type BaseFeeReceiptBSONUnmarshaler struct {
	Hint       string     `bson:"_hint"`
	CurrencyID CurrencyID `bson:"currency_id"`
	Amount     string     `bson:"total_fee"`
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
	r.currencyID = u.CurrencyID
	r.totalFee = u.Amount

	return nil
}

func (r FixedFeeReceipt) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":       r.Hint().String(),
			"currency_id": r.currencyID,
			"total_fee":   r.totalFee,
			"base_fee":    r.baseFee,
		},
	)
}

type FixedFeeReceiptBSONUnmarshaler struct {
	Hint       string     `bson:"_hint"`
	CurrencyID CurrencyID `bson:"currency_id"`
	TotalFee   string     `bson:"total_fee"`
	BaseFee    string     `bson:"base_fee"`
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
	r.currencyID = u.CurrencyID
	r.totalFee = u.TotalFee
	r.baseFee = u.BaseFee

	return nil
}

func (r FixedItemDataSizeExecutionFeeReceipt) MarshalBSON() ([]byte, error) {
	m := bson.M{
		"_hint":              r.Hint().String(),
		"currency_id":        r.currencyID,
		"total_fee":          r.totalFee,
		"base_fee":           r.baseFee,
		"item_unit_fee":      r.itemUnitFee,
		"item_count":         r.itemCount,
		"item_fee":           r.itemFee,
		"data_size_unit_fee": r.dataSizeUnitFee,
		"data_size_unit":     r.dataSizeUnit,
		"data_size":          r.dataSize,
		"data_size_fee":      r.dataSizeFee,
	}

	if r.executionCount != 0 {
		m["execution_count"] = r.executionCount
	}

	if r.executionUnitFee != "" {
		m["execution_unit_fee"] = r.executionUnitFee
	}

	if r.executionFee != "" {
		m["execution_fee"] = r.executionFee
	}

	return bsonenc.Marshal(m)
}

type FixedItemDataSizeExecutionFeeReceiptBSONUnmarshaler struct {
	Hint             string     `bson:"_hint"`
	CurrencyID       CurrencyID `bson:"currency_id"`
	TotalFee         string     `bson:"total_fee"`
	BaseFee          string     `bson:"base_fee"`
	ItemUnitFee      string     `bson:"item_unit_fee"`
	ItemCount        int        `bson:"item_count"`
	ItemFee          string     `bson:"item_fee"`
	DataSizeUnitFee  string     `bson:"data_size_unit_fee"`
	DataSizeUnit     int64      `bson:"data_size_unit"`
	DataSize         int        `bson:"data_size"`
	DataSizeFee      string     `bson:"data_size_fee"`
	ExecutionCount   int        `bson:"execution_count"`
	ExecutionUnitFee string     `bson:"execution_unit_fee"`
	ExecutionFee     string     `bson:"execution_fee"`
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
	r.currencyID = u.CurrencyID
	r.totalFee = u.TotalFee
	r.baseFee = u.BaseFee
	r.itemUnitFee = u.ItemUnitFee
	r.itemCount = u.ItemCount
	r.itemFee = u.ItemFee
	r.dataSizeUnitFee = u.DataSizeUnitFee
	r.dataSizeUnit = u.DataSizeUnit
	r.dataSize = u.DataSize
	r.dataSizeFee = u.DataSizeFee
	r.executionCount = u.ExecutionCount
	r.executionUnitFee = u.ExecutionUnitFee
	r.executionFee = u.ExecutionFee

	return nil
}

func (r CurrencyOperationReceipt) MarshalBSON() ([]byte, error) {
	m := bson.M{
		"_hint": r.Hint().String(),
	}

	if r.feeer != "" {
		m["feeer"] = r.feeer
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
	Feeer   string   `bson:"feeer"`
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
	r.feeer = u.Feeer
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
