package currency

import (
	"github.com/imfact-labs/currency-model/operation/test"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
)

type TestMintProcessor struct {
	*test.BaseTestOperationProcessorNoItem[Mint]
}

func NewTestMintProcessor(
	tp *test.TestProcessor,
) TestMintProcessor {
	t := test.NewBaseTestOperationProcessorNoItem[Mint](tp)

	return TestMintProcessor{&t}
}

func (t *TestMintProcessor) Create() *TestMintProcessor {
	t.Opr, _ = NewMintProcessor(base.MaxThreshold)(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)

	return t
}

func (t *TestMintProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool) *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestMintProcessor) SetAmount(am int64, cid types.CurrencyID, target []types.Amount) *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.SetAmount(am, cid, target)

	return t
}

func (t *TestMintProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestMintProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool) *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestMintProcessor) LoadOperation(fileName string) *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.LoadOperation(fileName)

	return t
}

func (t *TestMintProcessor) Print(fileName string) *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.Print(fileName)

	return t
}

func (t *TestMintProcessor) MakeOperation(receiver test.Account, amount types.Amount,
) *TestMintProcessor {
	//t.MockGetter.On("Get", mock.Anything).Return(nil, false, nil)

	op, _ := NewMint(NewMintFact([]byte("token"), receiver.Address(), amount))
	_ = op.NodeSign(t.NodePriv, t.NetworkID, t.NodeAddr)
	t.Op = op

	return t
}

func (t *TestMintProcessor) RunPreProcess() *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.RunPreProcess()

	return t
}

func (t *TestMintProcessor) RunProcess() *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.RunProcess()

	return t
}

func (t *TestMintProcessor) IsValid() *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.IsValid()

	return t
}

func (t *TestMintProcessor) Decode(fileName string) *TestMintProcessor {
	t.BaseTestOperationProcessorNoItem.Decode(fileName)

	return t
}
