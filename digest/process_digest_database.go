package digest

import (
	"context"

	mongodbst "github.com/imfact-labs/imfact-currency/digest/mongodb"
	"github.com/ProtoconNet/mitum2/isaac"
	isaacdatabase "github.com/ProtoconNet/mitum2/isaac/database"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/pkg/errors"
)

func ProcessDigesterDatabase(ctx context.Context) (context.Context, error) {
	var design YamlDigestDesign
	if err := util.LoadFromContext(ctx, ContextValueDigestDesign, &design); err != nil {
		return ctx, err
	}

	if design.Equal(YamlDigestDesign{}) || !design.Digest {
		return ctx, nil
	}

	conf := design.Database()

	switch {
	case conf.URI().Scheme == "mongodb", conf.URI().Scheme == "mongodb+srv":
		return processMongodbDatabase(ctx, design)
	default:
		return ctx, errors.Errorf("Unsupported database type, %v", conf.URI().Scheme)
	}
}

func processMongodbDatabase(ctx context.Context, design YamlDigestDesign) (context.Context, error) {
	conf := design.Database()

	/*
		ca, err := cache.NewCacheFromURI(conf.Cache().String())
		if err != nil {
			return ctx, err
		}
	*/

	var encs *encoder.Encoders
	if err := util.LoadFromContext(ctx, launch.EncodersContextKey, &encs); err != nil {
		return ctx, err
	}

	st, err := mongodbst.NewDatabaseFromURI(conf.URI().String(), encs)
	if err != nil {
		return ctx, err
	}

	if err := st.Initialize(); err != nil {
		return ctx, err
	}

	var db isaac.Database
	if err := util.LoadFromContextOK(ctx, launch.CenterDatabaseContextKey, &db); err != nil {
		return ctx, err
	}

	mst, ok := db.(*isaacdatabase.Center)
	if !ok {
		return ctx, errors.Errorf("expected isaacdatabase.Center, not %T", db)
	}

	dst, err := loadDigestDatabase(mst, st, false)
	if err != nil {
		return ctx, err
	}
	var log *logging.Logging
	if err := util.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, err
	}

	_ = dst.SetLogging(log)

	return context.WithValue(ctx, ContextValueDigestDatabase, dst), nil
}

func loadDigestDatabase(mst *isaacdatabase.Center, st *mongodbst.Database, readonly bool) (*Database, error) {
	var dst *Database
	if readonly {
		s, err := NewReadonlyDatabase(mst, st)
		if err != nil {
			return nil, err
		}
		dst = s
	} else {
		s, err := NewDatabase(mst, st)
		if err != nil {
			return nil, err
		}
		dst = s
	}

	if err := dst.Initialize(DefaultIndexes); err != nil {
		return nil, err
	}

	return dst, nil
}
