package contracts

import (
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
)

type ProposalOperationFactHintFunc func() func(hint.Hint) bool

type NewOperationProcessorInternalWithProposalFunc func(
	base.Height,
	base.ProposalSignFact,
	base.GetStateFunc,
) (base.OperationProcessor, error)

var (
	ProposalOperationFactHintContextKey = util.ContextKey("proposal-operation-fact-hint")
	OperationProcessorContextKey        = util.ContextKey("mitum-currency-operation-processor")
	OperationProcessorsMapBContextKey   = util.ContextKey("operation-processors-map-b")
)
