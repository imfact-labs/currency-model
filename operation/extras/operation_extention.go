package extras

import (
	"encoding/json"

	"github.com/btcsuite/btcutil/base58"
	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/state"
	didstate "github.com/imfact-labs/currency-model/state/did-registry"
	estate "github.com/imfact-labs/currency-model/state/extension"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

type Authentication interface {
	hint.Hinter
	util.IsValider
	util.Byter
	Contract() base.Address
	AuthenticationID() string
	ProofData() string
}

type Settlement interface {
	hint.Hinter
	util.IsValider
	util.Byter
	OpSender() base.Address
}

type ProxyPayer interface {
	hint.Hinter
	util.IsValider
	util.Byter
	ProxyPayer() base.Address
}

var BaseAuthenticationHint = hint.MustNewHint("mitum-extension-base-authentication-v0.0.1")
var AuthenticationExtensionType string = "authentication"

type BaseAuthentication struct {
	hint.BaseHinter
	contract         base.Address
	authenticationID string
	proofData        string
}

func NewBaseAuthentication(contract base.Address, authenticationID, proofData string) BaseAuthentication {
	return BaseAuthentication{
		BaseHinter:       hint.NewBaseHinter(BaseAuthenticationHint),
		contract:         contract,
		authenticationID: authenticationID,
		proofData:        proofData,
	}
}

func (ba BaseAuthentication) Contract() base.Address {
	return ba.contract
}

func (ba BaseAuthentication) AuthenticationID() string {
	return ba.authenticationID
}

func (ba BaseAuthentication) ProofData() string {
	return ba.proofData
}

func (ba BaseAuthentication) ExtType() string {
	return AuthenticationExtensionType
}

func (ba BaseAuthentication) Bytes() []byte {
	if ba.Equal(BaseAuthentication{}) {
		return []byte{}
	}
	var bs [][]byte
	bs = append(bs, ba.contract.Bytes())
	bs = append(bs, []byte(ba.authenticationID))
	bs = append(bs, []byte(ba.proofData))
	return util.ConcatBytesSlice(bs...)
}

func (ba BaseAuthentication) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, ba.contract); err != nil {
		return common.ErrValueInvalid.Wrap(err)
	}

	if len(ba.authenticationID) < 1 {
		return common.ErrValueInvalid.Wrap(errors.Errorf("empty authentication id"))
	}

	if len(ba.proofData) < 1 {
		return common.ErrValueInvalid.Wrap(errors.Errorf("empty proof data"))
	}

	return nil
}

func (ba BaseAuthentication) Equal(b BaseAuthentication) bool {
	if !ba.contract.Equal(b.contract) {
		return false
	}

	if ba.authenticationID != b.authenticationID {
		return false
	}

	if ba.proofData != b.proofData {
		return false
	}

	return true
}

func (ba BaseAuthentication) Verify(op base.Operation, getStateFunc base.GetStateFunc) error {
	var authentication types.VerificationRelationshipEntry
	var doc types.DIDDocument

	var factUser base.Address
	i, ok := op.Fact().(FactUser)
	if !ok {
		return common.ErrAccountNAth.Errorf("fact user not found")
	} else if factUser = i.FactUser(); factUser == nil {
		return common.ErrAccountNAth.Errorf("empty fact user")
	}

	authId, err := types.NewDIDURLRefFromString(ba.AuthenticationID())
	if err != nil {
		return err
	}

	if factUser.String() != authId.MethodSpecificID() {
		return common.ErrValueInvalid.Errorf("authentication id must be derived from the sender's DID")
	}

	if ba.Contract() == nil {
		return common.ErrValueInvalid.Errorf("empty contract address")
	}

	contract := ba.Contract()
	if st, err := state.ExistsState(didstate.DocumentStateKey(contract, authId.DID().String()), "did document", getStateFunc); err != nil {
		return common.ErrStateNF.Wrap(err)
	} else if doc, err = didstate.GetDocumentFromState(st); err != nil {
		return err
	}

	authentication, err = doc.Authentication(authId.String())
	if err != nil {
		return common.ErrValueInvalid.Wrap(err)
	}

	var iVrfMethod types.IVerificationMethod
	if authentication.Kind() == types.VMRefKindReference {
		iVrfMethod, err = doc.VerificationMethod(authId.String())
		if err != nil {
			return common.ErrValueInvalid.Wrap(err)
		}
	} else if authentication.Kind() == types.VMRefKindEmbedded {
		iVrfMethod = authentication.Method()
	} else {
		return common.ErrValueInvalid.Errorf("unknown authentication kind")
	}

	vrfMethod, ok := iVrfMethod.(types.VerificationMethod)
	if !ok {
		return errors.Errorf("expected VerificationMethod but %T", iVrfMethod)
	}

	switch vrfMethod.Type() {
	case types.AuthTypeECDSASECP, types.AuthTypeImFact:
		if vrfMethod.PublicKey() == nil {
			return common.ErrValueInvalid.Errorf("missing public key in EcdsaSecp256k1VerificationKey2019 type")
		}
		pubKey := vrfMethod.PublicKey()
		signature := base58.Decode(ba.ProofData())
		err = pubKey.Verify(op.Fact().Hash().Bytes(), signature)
		if err != nil {
			return common.ErrUserSignInvalid.Wrap(err)
		}
	case types.AuthTypeLinked:
		targetID := vrfMethod.TargetID()
		if targetID == nil {
			return common.ErrUserSignInvalid.Wrap(errors.Errorf("empty target ID in LinkedVerificationMethod type"))
		}

		var allowed []types.AllowedOperation
		switch t := op.Fact().(type) {
		case ActiveContract:
			for _, contract := range t.ActiveContract() {
				allowed = append(allowed, *types.NewAllowedOperation(contract, op.Hint()))
			}
		case ActiveContractOwnerHandlerOnly:
			for _, contract := range t.ActiveContractOwnerHandlerOnly() {
				allowed = append(allowed, *types.NewAllowedOperation(contract[0], op.Hint()))
			}
		case InActiveContractOwnerHandlerOnly:
			for _, contract := range t.InActiveContractOwnerHandlerOnly() {
				allowed = append(allowed, *types.NewAllowedOperation(contract[0], op.Hint()))
			}
		default:
			allowed = append(allowed, *types.NewAllowedOperation(nil, op.Hint()))
		}

		for _, allowedOp := range allowed {
			ok := vrfMethod.IsAllowed(allowedOp)
			if !ok {
				if allowedOp.Contract() == nil {
					return common.ErrValueInvalid.Errorf(
						"operation %s is not found in allowed operation", allowedOp.Operation().String())
				} else {
					return common.ErrValueInvalid.Errorf(
						"operation %s for contract %s is not found in allowed operation",
						allowedOp.Operation().String(), allowedOp.Contract().String(),
					)
				}
			}
		}

		var targetDoc types.DIDDocument
		if targetSt, err := state.ExistsState(didstate.DocumentStateKey(contract, targetID.DID().String()), "did document", getStateFunc); err != nil {
			return common.ErrStateNF.Wrap(err)
		} else if targetDoc, err = didstate.GetDocumentFromState(targetSt); err != nil {
			return err
		}
		tAuthentication, err := targetDoc.Authentication(targetID.String())
		if err != nil {
			return common.ErrValueInvalid.Wrap(err)
		}

		var iVrfMethod types.IVerificationMethod
		if tAuthentication.Kind() == types.VMRefKindReference {
			iVrfMethod, err = targetDoc.VerificationMethod(targetID.String())
			if err != nil {
				return common.ErrValueInvalid.Wrap(err)
			}
		} else if tAuthentication.Kind() == types.VMRefKindEmbedded {
			iVrfMethod = tAuthentication.Method()
		} else {
			return common.ErrValueInvalid.Errorf("unknown authentication kind")
		}

		vrfMethod, ok := iVrfMethod.(types.VerificationMethod)
		if !ok {
			return errors.Errorf("expected VerificationMethod but %T", iVrfMethod)
		}

		switch vrfMethod.Type() {
		case types.AuthTypeECDSASECP, types.AuthTypeImFact:
			if vrfMethod.PublicKey() == nil {
				return common.ErrValueInvalid.Errorf("missing public key in EcdsaSecp256k1VerificationKey2019 type")
			}
			pubKey := vrfMethod.PublicKey()
			signature := base58.Decode(ba.ProofData())
			err = pubKey.Verify(op.Fact().Hash().Bytes(), signature)
			if err != nil {
				return common.ErrUserSignInvalid.Wrap(err)
			}
		case types.AuthTypeLinked:
			return common.ErrValueInvalid.Errorf("target authentiation id should not point LinkedVerificationMethod type")
		}
	}

	return nil
}

var BaseSettlementHint = hint.MustNewHint("mitum-extension-base-settlement-v0.0.1")
var SettlementExtensionType string = "settlement"

type BaseSettlement struct {
	hint.BaseHinter
	opSender base.Address
}

func NewBaseSettlement(opSender base.Address) BaseSettlement {
	return BaseSettlement{
		BaseHinter: hint.NewBaseHinter(BaseSettlementHint),
		opSender:   opSender,
	}
}

func (bs BaseSettlement) OpSender() base.Address {
	return bs.opSender
}

func (bs BaseSettlement) Bytes() []byte {
	if bs.Equal(BaseSettlement{}) {
		return []byte{}
	}
	var b [][]byte
	b = append(b, bs.opSender.Bytes())
	return util.ConcatBytesSlice(b...)
}

func (bs BaseSettlement) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, bs.opSender); err != nil {
		return common.ErrValueInvalid.Wrap(err)
	}

	return nil
}

func (bs BaseSettlement) ExtType() string {
	return SettlementExtensionType
}

func (bs BaseSettlement) Verify(op base.Operation, getStateFunc base.GetStateFunc) error {
	opSender := bs.OpSender()
	if opSender == nil {
		return errors.Errorf("empty op sender")
	}

	if _, _, aErr, cErr := state.ExistsCAccount(opSender, "op sender", true, false, getStateFunc); aErr != nil {
		return aErr
	} else if cErr != nil {
		return common.ErrPreProcess.
			Wrap(common.ErrCAccountNA.
				Errorf("%v", cErr))
	}

	if err := state.CheckFactSignsByState(opSender, op.Signs(), getStateFunc); err != nil {
		return err
	}

	return nil
}

func (bs BaseSettlement) Equal(b BaseSettlement) bool {
	if !bs.opSender.Equal(b.opSender) {
		return false
	}

	return true
}

var BaseProxyPayerHint = hint.MustNewHint("mitum-extension-base-proxy-payer-v0.0.1")
var ProxyPayerExtensionType string = "proxy_payer"

type BaseProxyPayer struct {
	hint.BaseHinter
	proxyPayer base.Address
}

func NewBaseProxyPayer(proxyPayer base.Address) BaseProxyPayer {
	return BaseProxyPayer{
		BaseHinter: hint.NewBaseHinter(BaseProxyPayerHint),
		proxyPayer: proxyPayer,
	}
}

func (bs BaseProxyPayer) ProxyPayer() base.Address {
	return bs.proxyPayer
}

func (bs BaseProxyPayer) Bytes() []byte {
	if bs.Equal(BaseProxyPayer{}) {
		return []byte{}
	}
	var b [][]byte
	if bs.proxyPayer != nil {
		b = append(b, bs.proxyPayer.Bytes())
	}
	return util.ConcatBytesSlice(b...)
}

func (bs BaseProxyPayer) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, bs.proxyPayer); err != nil {
		return common.ErrValueInvalid.Wrap(err)
	}

	return nil
}

func (bs BaseProxyPayer) ExtType() string {
	return ProxyPayerExtensionType
}

func (bs BaseProxyPayer) Verify(op base.Operation, getStateFunc base.GetStateFunc) error {
	proxyPayer := bs.ProxyPayer()
	if proxyPayer == nil {
		return common.ErrValueInvalid.Errorf("empty proxy payer")
	}
	feeBaser, ok := op.Fact().(FeeAble)
	if !ok {
		return common.ErrTypeMismatch.Errorf("expected FeeAble but %T", op.Fact())
	}

	sender := feeBaser.FeePayer()
	if sender == nil {
		return common.ErrValueInvalid.Errorf("empty fact sender")
	}

	if _, cSt, aErr, cErr := state.ExistsCAccount(proxyPayer, "proxy payer", true, true, getStateFunc); aErr != nil {
		return aErr
	} else if cErr != nil {
		return errors.Errorf("%v", cErr)
	} else if ca, err := estate.LoadCAStateValue(cSt); err != nil {
		return err
	} else if !ca.IsRecipients(sender) {
		return common.ErrAccountNAth.Errorf("user, %v is not recipient of proxy payer, %v", sender, proxyPayer)
	}

	return nil
}

func (bs BaseProxyPayer) Equal(b BaseProxyPayer) bool {
	if !bs.proxyPayer.Equal(b.proxyPayer) {
		return false
	}

	return true
}

type ExtendedOperation struct {
	common.BaseOperation
	*BaseOperationExtensions
}

func NewExtendedOperation(hint hint.Hint, fact base.Fact) ExtendedOperation {
	return ExtendedOperation{
		BaseOperation:           common.NewBaseOperation(hint, fact),
		BaseOperationExtensions: NewBaseOperationExtensions(),
	}
}

func (op ExtendedOperation) IsValid(networkID []byte) error {
	if err := op.BaseOperation.IsValid(networkID); err != nil {
		return err
	}
	if err := op.BaseOperationExtensions.IsValid(networkID); err != nil {
		return err
	}

	return nil
}

func (op *ExtendedOperation) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}

	op.SetHash(op.hash())

	return nil
}

func (op ExtendedOperation) hash() util.Hash {
	return valuehash.NewSHA256(op.HashBytes())
}

func (op ExtendedOperation) HashBytes() []byte {
	var bs [][]byte
	bs = append(bs, op.BaseOperation.HashBytes())

	if op.BaseOperationExtensions != nil {
		bs = append(bs, op.BaseOperationExtensions.Bytes())
	}

	return util.ConcatBytesSlice(bs...)
}

type OperationExtension interface {
	ExtType() string
	Verify(base.Operation, base.GetStateFunc) error
	util.IsValider
	util.Byter
}

type OperationExtensions interface {
	util.IsValider
	util.Byter
	Verify(base.Operation, base.GetStateFunc) error
	Extension(string) OperationExtension
	Extensions() map[string]OperationExtension
	AddExtension(OperationExtension) error
}

type BaseOperationExtensions struct {
	extension map[string]OperationExtension
}

func NewBaseOperationExtensions() *BaseOperationExtensions {
	return &BaseOperationExtensions{
		extension: make(map[string]OperationExtension),
	}

}

func (be BaseOperationExtensions) Bytes() []byte {
	var bs [][]byte
	if be.extension != nil {
		extension, _ := json.Marshal(be.extension)
		bs = append(bs, valuehash.NewSHA256(extension).Bytes())
	}

	return util.ConcatBytesSlice(bs...)
}

func (be BaseOperationExtensions) Verify(op base.Operation, getStateFunc base.GetStateFunc) error {
	auth := be.Extension(AuthenticationExtensionType)
	if auth != nil {
		if err := auth.IsValid(nil); err != nil {
			return err
		}

		if err := auth.Verify(op, getStateFunc); err != nil {
			return err
		}
	}
	settlement := be.Extension(SettlementExtensionType)
	if settlement != nil {
		if err := settlement.IsValid(nil); err != nil {
			return err
		}

		if err := settlement.Verify(op, getStateFunc); err != nil {
			return err
		}
	}
	proxyPayer := be.Extension(ProxyPayerExtensionType)
	if proxyPayer != nil {
		if err := proxyPayer.IsValid(nil); err != nil {
			return err
		}

		if err := proxyPayer.Verify(op, getStateFunc); err != nil {
			return err
		}
	}

	return nil
}

func (be BaseOperationExtensions) IsValid(networkID []byte) error {
	for _, ext := range be.extension {
		if err := ext.IsValid(networkID); err != nil {
			return err
		}
	}
	return nil
}

func (be BaseOperationExtensions) Extension(extType string) OperationExtension {
	if len(be.extension) < 1 {
		return nil
	}

	extension, ok := be.extension[extType]
	if !ok {
		return nil
	}

	return extension
}

func (be BaseOperationExtensions) Extensions() map[string]OperationExtension {
	return be.extension
}

func (be *BaseOperationExtensions) AddExtension(extension OperationExtension) error {
	if err := util.CheckIsValiders(nil, false, extension); err != nil {
		return err
	}

	_, ok := be.extension[extension.ExtType()]
	if ok {
		return errors.Errorf("%s is already added", extension.ExtType())
	}

	be.extension[extension.ExtType()] = extension

	return nil
}

const (
	DuplicationKeyTypeNewAddress       types.DuplicationKeyType = "new-address"
	DuplicationKeyTypeSender           types.DuplicationKeyType = "currency-sender"
	DuplicationKeyTypeCurrency         types.DuplicationKeyType = "currency-id"
	DuplicationKeyTypeNewContract      types.DuplicationKeyType = "new-contract"
	DuplicationKeyTypeContractStatus   types.DuplicationKeyType = "contract-status"
	DuplicationKeyTypeContractWithdraw types.DuplicationKeyType = "contract-withdraw"
	DuplicationKeyTypeDIDAccount       types.DuplicationKeyType = "did-account"
)

type DeDupeKeyer interface {
	DupKey() (map[types.DuplicationKeyType][]string, error)
}

// FeeAble is an interface type for fee calculation. Operations than requires fee must implement this interface.
type FeeAble interface {
	FeeBase() map[types.CurrencyID][]common.Big
	FeeItemCount() (uint, bool)
	FeePayer() base.Address
}

const (
	ZeroItem  = 0
	HasItem   = true
	HasNoItem = false
)

// VerifyFeeAble function checks existence of currency id
func VerifyFeeAble(fact FeeAble, getStateFunc base.GetStateFunc) base.OperationProcessReasonError {
	if len(fact.FeeBase()) < 1 {
		return base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("fail to get Fee Base, empty Fee Base "))
	}

	for cid := range fact.FeeBase() {
		_, err := state.ExistsCurrencyPolicy(cid, getStateFunc)
		if err != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", err))
		}
	}

	return nil
}

// FactUser is an interface type for finding the user associated with User Operation
type FactUser interface {
	FactUser() base.Address
}

// VerifyFactUser function checks
// existence of user account
// it is not a contract account
func VerifyFactUser(fact FactUser, getStateFunc base.GetStateFunc) base.OperationProcessReasonError {
	user := fact.FactUser()

	if user == nil {
		return base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("failed to get FactUser, empty user account"))
	}
	if _, _, aErr, cErr := state.ExistsCAccount(user, "sender", true, false, getStateFunc); aErr != nil {
		return base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Errorf("%v", aErr))
	} else if cErr != nil {
		return base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).Errorf("%v", cErr))
	}

	return nil
}

// ContractOwnerOnly is an interface type for operations that must be controlled by contract owner
// Withdraw, UpdateHandler, UpdateRecipient
type ContractOwnerOnly interface {
	ContractOwnerOnly() [][2]base.Address // contract, sender
}

// VerifyContractOwnerOnly function checks
// existence of contract account
// sender is owner of contract account
func VerifyContractOwnerOnly(fact ContractOwnerOnly, getStateFunc base.GetStateFunc) base.OperationProcessReasonError {
	for _, addresses := range fact.ContractOwnerOnly() {
		contract := addresses[0]
		sender := addresses[1]
		if contract == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("failed to get contract account, empty contract account"))
		}
		if sender == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("failed to get sender, empty sender account"))
		}

		if _, cSt, aErr, cErr := state.ExistsCAccount(contract, "contract", true, true, getStateFunc); aErr != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", aErr))
		} else if cErr != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", cErr))
		} else if status, err := estate.StateContractAccountValue(cSt); err != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMStateValInvalid).
					Errorf("%v", cErr))
		} else if !status.Owner().Equal(sender) {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMAccountNAth).
					Errorf("sender %v is not owner of contract account", sender))
		}
	}

	return nil
}

// InActiveContractOwnerHandlerOnly is an interface type for operations that activate an inactive contract
// and must be authorized by owner or handler (e.g., RegisterModel)
type InActiveContractOwnerHandlerOnly interface {
	InActiveContractOwnerHandlerOnly() [][2]base.Address // contract, sender
}

// VerifyInActiveContractOwnerHandlerOnly function checks existence of contract account
// sender is owner of contract account inactive contract account
func VerifyInActiveContractOwnerHandlerOnly(fact InActiveContractOwnerHandlerOnly, getStateFunc base.GetStateFunc) base.OperationProcessReasonError {
	for _, addresses := range fact.InActiveContractOwnerHandlerOnly() {
		contract := addresses[0]
		sender := addresses[1]

		if contract == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("failed to get contract account, empty contract account"))
		}
		if sender == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("failed to get sender, empty sender account"))
		}

		_, cSt, aErr, cErr := state.ExistsCAccount(contract, "contract", true, true, getStateFunc)
		if aErr != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", aErr))
		} else if cErr != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", cErr))
		}

		ca, err := estate.CheckCAAuthFromState(cSt, sender)
		if err != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", err))
		}

		if ca == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMValueInvalid).Errorf(
					"contract account value is nil"))
		}

		if ca.IsActive() {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMServiceE).Errorf(
					"contract account %v has already been activated", contract))
		}
	}

	return nil
}

// ActiveContractOwnerHandlerOnly is an interface type for operations on an activated contract
// that must be authorized by owner or handler(e.g., Mint)
type ActiveContractOwnerHandlerOnly interface {
	ActiveContractOwnerHandlerOnly() [][2]base.Address // contract, sender
}

// VerifyActiveContractOwnerHandlerOnly function checks
// existence of contract account
// sender is owner of contract account
func VerifyActiveContractOwnerHandlerOnly(fact ActiveContractOwnerHandlerOnly, getStateFunc base.GetStateFunc) base.OperationProcessReasonError {
	for _, addresses := range fact.ActiveContractOwnerHandlerOnly() {
		contract := addresses[0]
		sender := addresses[1]

		if contract == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("failed to get contract account, empty contract account"))
		}
		if sender == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("failed to get sender, empty sender account"))
		}

		_, cSt, aErr, cErr := state.ExistsCAccount(contract, "contract", true, true, getStateFunc)
		if aErr != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", aErr))
		} else if cErr != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", cErr))
		}

		ca, err := estate.CheckCAAuthFromState(cSt, sender)
		if err != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", err))
		}

		if ca == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMValueInvalid).Errorf(
					"contract account value is nil"))
		}

		if !ca.IsActive() {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMServiceNF).Errorf(
					"contract account %v has not been activated", contract))
		}
	}

	return nil
}

// ActiveContract is an interface type for operations on an active contract(e.g., CreateDID)
type ActiveContract interface {
	ActiveContract() []base.Address
}

// VerifyActiveContract function checks
// existence of contract account
// active contract account
func VerifyActiveContract(fact ActiveContract, getStateFunc base.GetStateFunc) base.OperationProcessReasonError {
	for _, contract := range fact.ActiveContract() {
		if contract == nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("failed to get contract account, empty contract account"))
		}

		_, cSt, aErr, cErr := state.ExistsCAccount(contract, "contract", true, true, getStateFunc)
		if aErr != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", aErr))
		} else if cErr != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", cErr))
		}

		ca, err := estate.LoadCAStateValue(cSt)
		if err != nil {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", err))
		}

		if !ca.IsActive() {
			return base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMServiceNF).Errorf(
					"contract account %v has not been activated", contract))
		}
	}

	return nil
}
