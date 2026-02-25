package module

import (
	"github.com/imfact-labs/currency-model/api"
	"github.com/imfact-labs/currency-model/app/modulekit"
	"github.com/imfact-labs/currency-model/app/runtime/spec"
	"github.com/imfact-labs/currency-model/app/runtime/steps"
)

const ID = "currency"

type Module struct{}

var _ modulekit.ModelModule = Module{}

func (Module) ID() string {
	return ID
}

func (Module) Register(reg *modulekit.Registry) error {
	if err := reg.AddHinters(ID, spec.Hinters...); err != nil {
		return err
	}

	if err := reg.AddSupportedFacts(ID, spec.SupportedProposalOperationFactHinters...); err != nil {
		return err
	}

	if err := reg.AddOperationProcessors(ID, modulekit.OperationProcessors{
		Name:      steps.PNameOperationProcessorsMap,
		Func:      steps.POperationProcessorsMap,
		SupportsA: true,
		SupportsB: true,
	}); err != nil {
		return err
	}

	if err := reg.AddAPIRoutes(
		ID,
		modulekit.APIRoute{Path: api.HandlerPathNodeInfo, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathNodeInfoProm, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathNodeMetric, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathNodeMetricProm, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathSend, Methods: []string{"POST"}},
		modulekit.APIRoute{Path: api.HandlerPathQueueSend, Methods: []string{"POST"}},
		modulekit.APIRoute{Path: api.HandlerPathCurrencies, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathCurrency, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathManifests, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathOperations, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathOperationsByHash, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathOperation, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathOperationsByHeight, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathManifestByHeight, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathManifestByHash, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathBlockByHeight, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathBlockByHash, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathAccount, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathAccountOperations, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathAccounts, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathDIDDesign, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathDIDData, Methods: []string{"GET"}},
		modulekit.APIRoute{Path: api.HandlerPathDIDDocument, Methods: []string{"GET"}},
	); err != nil {
		return err
	}

	if err := reg.AddAPIHandlers(ID, modulekit.APIHandlerInitializer{
		Key: "currency.api.handlers",
		Register: func(hd *api.Handlers, digestEnabled bool) {
			api.SetHandlers(hd, digestEnabled)
		},
	}); err != nil {
		return err
	}

	return reg.AddCLICommands(
		ID,
		modulekit.CLICommand{Key: "operation.currency", Description: "currency operation"},
		modulekit.CLICommand{Key: "operation.suffrage", Description: "suffrage operation"},
		modulekit.CLICommand{Key: "operation.did", Description: "did operation"},
	)
}
