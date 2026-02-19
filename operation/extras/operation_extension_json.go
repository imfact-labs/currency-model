package extras

import (
	"encoding/json"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

type BaseAuthenticationJSONMarshaler struct {
	hint.BaseHinter
	Contract         base.Address `json:"contract"`
	DID              string       `json:"did"`
	AuthenticationID string       `json:"authentication_id"`
	ProofData        string       `json:"proof_data"`
}

func (ba BaseAuthentication) JSONMarshaler() BaseAuthenticationJSONMarshaler {
	return BaseAuthenticationJSONMarshaler{
		BaseHinter:       ba.BaseHinter,
		Contract:         ba.contract,
		AuthenticationID: ba.authenticationID,
		ProofData:        ba.proofData,
	}
}

func (ba BaseAuthentication) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(ba.JSONMarshaler())
}

type BaseAuthenticationJSONUnmarshaler struct {
	Hint             hint.Hint `json:"_hint"`
	Contract         string    `json:"contract"`
	AuthenticationID string    `json:"authentication_id"`
	ProofData        string    `json:"proof_data"`
}

func (ba *BaseAuthentication) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u BaseAuthenticationJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *ba)
	}

	ba.BaseHinter = hint.NewBaseHinter(u.Hint)
	a, err := base.DecodeAddress(u.Contract, enc)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *ba)
	}
	ba.contract = a
	ba.authenticationID = u.AuthenticationID
	ba.proofData = u.ProofData

	return nil
}

type BaseSettlementJSONMarshaler struct {
	hint.BaseHinter
	OpSender base.Address `json:"op_sender"`
}

func (bs BaseSettlement) JSONMarshaler() BaseSettlementJSONMarshaler {
	return BaseSettlementJSONMarshaler{
		BaseHinter: bs.BaseHinter,
		OpSender:   bs.opSender,
	}
}

func (bs BaseSettlement) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(bs.JSONMarshaler())
}

type BaseSettlementJSONUnmarshaler struct {
	Hint     hint.Hint `json:"_hint"`
	OpSender string    `json:"op_sender"`
}

func (bs *BaseSettlement) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u BaseSettlementJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *bs)
	}

	bs.BaseHinter = hint.NewBaseHinter(u.Hint)
	a, err := base.DecodeAddress(u.OpSender, enc)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *bs)
	}
	bs.opSender = a

	return nil
}

type BaseProxyPayerJSONMarshaler struct {
	hint.BaseHinter
	ProxyPayer base.Address `json:"proxy_payer"`
}

func (bs BaseProxyPayer) JSONMarshaler() BaseProxyPayerJSONMarshaler {
	return BaseProxyPayerJSONMarshaler{
		BaseHinter: bs.BaseHinter,
		ProxyPayer: bs.proxyPayer,
	}
}

func (bs BaseProxyPayer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(bs.JSONMarshaler())
}

type BaseProxyPayerJSONUnmarshaler struct {
	Hint       hint.Hint `json:"_hint"`
	ProxyPayer string    `json:"proxy_payer"`
}

func (bs *BaseProxyPayer) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u BaseProxyPayerJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *bs)
	}

	bs.BaseHinter = hint.NewBaseHinter(u.Hint)
	a, err := base.DecodeAddress(u.ProxyPayer, enc)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *bs)
	}
	bs.proxyPayer = a

	return nil
}

type BaseOperationExtensionsJSONMarshaler struct {
	Extension map[string]OperationExtension `json:"extension"`
}

func (be BaseOperationExtensions) JSONMarshaler() BaseOperationExtensionsJSONMarshaler {
	return BaseOperationExtensionsJSONMarshaler{
		Extension: be.extension,
	}
}

func (be BaseOperationExtensions) MarshalJSON() ([]byte, error) {
	if len(be.extension) < 1 {
		return []byte{}, nil
	}
	return util.MarshalJSON(BaseOperationExtensionsJSONMarshaler{
		Extension: be.extension,
	})
}

type BaseOperationExtensionsJSONUnmarshaler struct {
	Extension json.RawMessage
}

func (be *BaseOperationExtensions) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u BaseOperationExtensionsJSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *be)
	}

	extensions := make(map[string]OperationExtension)
	m, err := enc.DecodeMap(u.Extension)
	if err != nil {
		return err
	}
	for k, v := range m {
		extension, ok := v.(OperationExtension)
		if !ok {
			return errors.Errorf("expected OperationExtension, not %T", v)
		}

		extensions[k] = extension
	}
	be.extension = extensions

	return nil
}
