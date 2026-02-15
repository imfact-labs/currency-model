module github.com/imfact-labs/imfact-currency

go 1.24.0

toolchain go1.24.6

require (
	github.com/ProtoconNet/mitum2 v0.0.0-20251029064552-48aef1088f5f
	github.com/btcsuite/btcd/btcec/v2 v2.3.5
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/ethereum/go-ethereum v1.14.13
	github.com/json-iterator/go v1.1.12
	github.com/multiformats/go-multibase v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.34.0
	go.mongodb.org/mongo-driver/v2 v2.5.0
	golang.org/x/crypto v0.45.0
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/beevik/ntp v1.4.3 // indirect
	github.com/bluele/gcache v0.0.2 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic v1.14.1 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/holiman/uint256 v1.3.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mr-tron/base58 v1.1.0 // indirect
	github.com/multiformats/go-base32 v0.0.3 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/quic-go/quic-go v0.54.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.uber.org/mock v0.5.0 // indirect
	golang.org/x/arch v0.7.0 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hashicorp/memberlist => github.com/spikeekips/memberlist v0.0.0-20230626195851-39f17fa10d23 // latest fix-data-race branch
