//go:build dev
// +build dev

package cmds

import (
	"context"
	"net/http"

	"github.com/imfact-labs/imfact-currency/api"
	"github.com/imfact-labs/imfact-currency/digest"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
)

func (cmd *RunCommand) pDigestAPIHandlers(ctx context.Context) (context.Context, error) {
	var params *launch.LocalParams
	var local base.LocalNode
	var design digest.YamlDigestDesign

	if err := util.LoadFromContextOK(ctx,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
		digest.ContextValueDigestDesign, &design,
	); err != nil {
		return nil, err
	}

	if design.Equal(digest.YamlDigestDesign{}) {
		return ctx, nil
	}

	cache, err := LoadCache(cmd.log, ctx, design)
	if err != nil {
		return ctx, err
	}

	var dnt *api.HTTP2Server
	if err := util.LoadFromContext(ctx, api.ContextValueDigestNetwork, &dnt); err != nil {
		return ctx, err
	}

	router := dnt.Router()
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	handlers, err := SetDigestAPIDefaultHandlers(cmd.log, ctx, params, cache, router, dnt.Queue())
	if err != nil {
		return ctx, err
	}

	if err := handlers.Initialize(); err != nil {
		return ctx, err
	}
	api.SetHandlers(handlers, design.Digest)

	return ctx, nil
}
