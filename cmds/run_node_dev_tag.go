//go:build dev
// +build dev

package cmds

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"net/http"
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

	cache, err := cmd.loadCache(ctx, design)
	if err != nil {
		return ctx, err
	}

	var dnt *digest.HTTP2Server
	if err := util.LoadFromContext(ctx, digest.ContextValueDigestNetwork, &dnt); err != nil {
		return ctx, err
	}

	router := dnt.Router()
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	handlers, err := cmd.setDigestAPIDefaultHandlers(ctx, params, cache, router, dnt.Queue())
	if err != nil {
		return ctx, err
	}

	if err := handlers.Initialize(design.Digest); err != nil {
		return ctx, err
	}

	return ctx, nil
}
