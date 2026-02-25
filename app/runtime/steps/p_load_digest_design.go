package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/mitum2/util/ps"
	"gopkg.in/yaml.v3"
)

var PNameDigestDesign = ps.Name("digest-design")

func PLoadDigestDesign(pctx context.Context) (context.Context, error) {
	e := util.StringError("load design")

	var log *logging.Logging
	var flag launch.DesignFlag

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.DesignFlagContextKey, &flag,
	); err != nil {
		return pctx, e.Wrap(err)
	}

	switch flag.Scheme() {
	case "file":
		b, err := os.ReadFile(filepath.Clean(flag.URL().Path))
		if err != nil {
			return pctx, e.Wrap(err)
		}

		var m struct {
			API *digest.YamlDigestDesign `yaml:"api"`
			//Sequencer *digest.YamlSequencerDesign `yaml:"sequencer"`
		}

		nb, err := util.ReplaceEnvVariables(b)
		if err != nil {
			return pctx, e.Wrap(err)
		}

		if err := yaml.Unmarshal(nb, &m); err != nil {
			return pctx, e.Wrap(err)
		} else if m.API != nil {
			if i, err := m.API.Set(pctx); err != nil {
				return pctx, e.Wrap(err)
			} else {
				pctx = i
			}
		}

		if m.API == nil {
			pctx = context.WithValue(pctx, digest.ContextValueDigestDesign, digest.YamlDigestDesign{})
		} else {
			pctx = context.WithValue(pctx, digest.ContextValueDigestDesign, *m.API)
			log.Log().Debug().Object("design", *m.API).Msg("digest design loaded")
		}
		//pctx = context.WithValue(pctx, digest.ContextValueSequencerDesign, *m.Sequencer)
	default:
		return pctx, e.Errorf("Unknown digest design uri, %q", flag.URL())
	}

	return pctx, nil
}
