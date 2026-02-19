package digest

import (
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/ps"
)

var (
	ContextValueDigestDesign    util.ContextKey = "digest_design"
	ContextValueSequencerDesign util.ContextKey = "sequencer_design"
	ContextValueDigestDatabase  util.ContextKey = "digest_database"

	ContextValueDigester     util.ContextKey = "digester"
	ContextValueLocalNetwork util.ContextKey = "local_network"
)

var (
	PNameDigesterDataBase = ps.Name("digester_database")
)
