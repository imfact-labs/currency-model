package steps

import (
	"context"

	"github.com/imfact-labs/currency-model/operation/isaac"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/pkg/errors"
)

func PSuffrageCandidateLimiterSet(pctx context.Context) (context.Context, error) {
	e := util.StringError("prepare SuffrageCandidateLimiterSet")

	var db isaac.Database
	if err := util.LoadFromContextOK(pctx, launch.CenterDatabaseContextKey, &db); err != nil {
		return pctx, e.Wrap(err)
	}

	set := hint.NewCompatibleSet[base.SuffrageCandidateLimiterFunc](8) //nolint:gomnd //...

	if err := set.Add(
		isaacoperation.FixedSuffrageCandidateLimiterRuleHint,
		base.SuffrageCandidateLimiterFunc(FixedSuffrageCandidateLimiterFunc()),
	); err != nil {
		return pctx, e.Wrap(err)
	}

	if err := set.Add(
		isaacoperation.MajoritySuffrageCandidateLimiterRuleHint,
		base.SuffrageCandidateLimiterFunc(MajoritySuffrageCandidateLimiterFunc(db)),
	); err != nil {
		return pctx, e.Wrap(err)
	}

	return context.WithValue(pctx, launch.SuffrageCandidateLimiterSetContextKey, set), nil
}

func MajoritySuffrageCandidateLimiterFunc(
	db isaac.Database,
) func(base.SuffrageCandidateLimiterRule) (base.SuffrageCandidateLimiter, error) {
	return func(rule base.SuffrageCandidateLimiterRule) (base.SuffrageCandidateLimiter, error) {
		var i isaacoperation.MajoritySuffrageCandidateLimiterRule
		if err := util.SetInterfaceValue(rule, &i); err != nil {
			return nil, err
		}

		proof, found, err := db.LastSuffrageProof()

		switch {
		case err != nil:
			return nil, errors.WithMessagef(err, "get last suffrage for MajoritySuffrageCandidateLimiter")
		case !found:
			return nil, errors.Errorf("last suffrage not found for MajoritySuffrageCandidateLimiter")
		}

		suf, err := proof.Suffrage()
		if err != nil {
			return nil, errors.WithMessagef(err, "get suffrage for MajoritySuffrageCandidateLimiter")
		}

		return isaacoperation.NewMajoritySuffrageCandidateLimiter(
			i,
			func() (uint64, error) {
				return uint64(suf.Len()), nil
			},
		), nil
	}
}

func FixedSuffrageCandidateLimiterFunc() func(
	base.SuffrageCandidateLimiterRule,
) (base.SuffrageCandidateLimiter, error) {
	return func(rule base.SuffrageCandidateLimiterRule) (base.SuffrageCandidateLimiter, error) {
		switch i, err := util.AssertInterfaceValue[isaacoperation.FixedSuffrageCandidateLimiterRule](rule); {
		case err != nil:
			return nil, err
		default:
			return isaacoperation.NewFixedSuffrageCandidateLimiter(i), nil
		}
	}
}
