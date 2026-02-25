package steps

import (
	"context"

	"github.com/imfact-labs/currency-model/app/runtime/contracts"
	"github.com/imfact-labs/currency-model/operation/currency"
	did "github.com/imfact-labs/currency-model/operation/did-registry"
	"github.com/imfact-labs/currency-model/operation/extension"
	isaacoperation "github.com/imfact-labs/currency-model/operation/isaac"
	"github.com/imfact-labs/currency-model/operation/processor"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
)

var (
	PNameOperationProcessorsMap = launch.PNameOperationProcessorsMap
)

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return pctx, err
	}

	limiterF, err := launch.NewSuffrageCandidateLimiterFunc(pctx)
	if err != nil {
		return pctx, err
	}

	setA := hint.NewCompatibleSet[isaac.NewOperationProcessorInternalFunc](1 << 9)
	setB := hint.NewCompatibleSet[contracts.NewOperationProcessorInternalWithProposalFunc](1 << 9)

	opr := processor.NewOperationProcessor()
	err = opr.SetCheckDuplicationFunc(processor.CheckDuplication)
	if err != nil {
		return pctx, err
	}
	err = opr.SetGetNewProcessorFunc(processor.GetNewProcessor)
	if err != nil {
		return pctx, err
	}
	if err := opr.SetProcessor(
		currency.CreateAccountHint,
		currency.NewCreateAccountProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.UpdateKeyHint,
		currency.NewUpdateKeyProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.TransferHint,
		currency.NewTransferProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.RegisterCurrencyHint,
		currency.NewRegisterCurrencyProcessor(isaacParams.Threshold()),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.UpdateCurrencyHint,
		currency.NewUpdateCurrencyProcessor(isaacParams.Threshold()),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		currency.MintHint,
		currency.NewMintProcessor(isaacParams.Threshold()),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		extension.CreateContractAccountHint,
		extension.NewCreateContractAccountProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		extension.UpdateHandlerHint,
		extension.NewUpdateHandlerProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		extension.UpdateRecipientHint,
		extension.NewUpdateRecipientProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		extension.WithdrawHint,
		extension.NewWithdrawProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		did.RegisterModelHint,
		did.NewRegisterModelProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		did.CreateDIDHint,
		did.NewCreateDIDProcessor(),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessor(
		did.UpdateDIDDocumentHint,
		did.NewUpdateDIDDocumentProcessor(),
	); err != nil {
		return pctx, err
	}

	_ = setA.Add(currency.CreateAccountHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(currency.UpdateKeyHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(currency.TransferHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(currency.RegisterCurrencyHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(currency.UpdateCurrencyHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(currency.MintHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(extension.CreateContractAccountHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(extension.UpdateHandlerHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(extension.UpdateRecipientHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(extension.WithdrawHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(did.CreateDIDHint, func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
		return opr.New(
			height,
			getStatef,
			nil,
			nil,
		)
	})

	_ = setA.Add(did.RegisterModelHint, func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
		return opr.New(
			height,
			getStatef,
			nil,
			nil,
		)
	})

	_ = setA.Add(did.UpdateDIDDocumentHint, func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
		return opr.New(
			height,
			getStatef,
			nil,
			nil,
		)
	})

	_ = setA.Add(isaacoperation.SuffrageCandidateHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			policy := db.LastNetworkPolicy()
			if policy == nil { // NOTE Usually it means empty block data
				return nil, nil
			}

			return isaacoperation.NewSuffrageCandidateProcessor(
				height,
				getStatef,
				limiterF,
				nil,
				policy.SuffrageCandidateLifespan(),
			)
		})

	_ = setA.Add(isaacoperation.SuffrageJoinHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			policy := db.LastNetworkPolicy()
			if policy == nil { // NOTE Usually it means empty block data
				return nil, nil
			}

			return isaacoperation.NewSuffrageJoinProcessor(
				height,
				isaacParams.Threshold(),
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(isaac.SuffrageExpelOperationHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			policy := db.LastNetworkPolicy()
			if policy == nil { // NOTE Usually it means empty block data
				return nil, nil
			}

			return isaacoperation.NewSuffrageExpelProcessor(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(isaacoperation.SuffrageDisjoinHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return isaacoperation.NewSuffrageDisjoinProcessor(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setA.Add(isaacoperation.NetworkPolicyHint,
		func(height base.Height, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			return isaacoperation.NewNetworkPolicyProcessor(
				height,
				isaacParams.Threshold(),
				getStatef,
				nil,
				nil,
			)
		})

	//var f ProposalOperationFactHintFunc = IsSupportedProposalOperationFactHintFunc

	pctx = context.WithValue(pctx, contracts.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, setA)     //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, contracts.OperationProcessorsMapBContextKey, setB) //revive:disable-line:modifies-parameter
	//pctx = context.WithValue(pctx, ProposalOperationFactHintContextKey, f)

	return pctx, nil
}
