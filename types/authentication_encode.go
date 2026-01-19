package types

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (v *BaseVerificationMethod) unpack(
	id, vrfType, controller string,
) error {
	did, err := NewDIDURLRefFromString(id)
	if err != nil {
		return err
	}
	v.id = *did
	v.verificationType = VerificationMethodType(vrfType)
	cont, err := NewDIDRefFromString(controller)
	if err != nil {
		return err
	}
	v.controller = *cont

	return nil
}

func (v *VerificationMethod) unpack(
	authtype string, pubKeyJwk *JWK, pubKeyMultibase, pubKey string, tid string, allowed []AllowedOperation,
) error {
	if pubKey != "" {
		pbKey, err := ParseMEPublickey(pubKey)
		if err != nil {
			return err
		}
		v.publicKey = pbKey
	}
	v.publicKeyMultibase = pubKeyMultibase
	switch authtype {
	case "EcdsaSecp256k1VerificationKey2019":
		if pubKeyMultibase == "" {
			if pubKey == "" {
				return errors.New("invalid EcdsaSecp256k1VerificationKey2019 type")
			} else {
				err := v.SetPublicKeyMultibase(v.publicKey)
				if err != nil {
					return err
				}
			}
		}
	case "EcdsaSecp256k1VerificationKeyImFact2025":
		if pubKey == "" {
			return errors.New("invalid EcdsaSecp256k1VerificationKeyImFact2025 type")
		}
	}

	v.publicKeyJwk = pubKeyJwk

	if tid != "" {
		targetID, err := NewDIDURLRefFromString(tid)
		if err != nil {
			return err
		}
		v.targetID = targetID
	}

	v.allowed = allowed

	return nil
}

func (a *AllowedOperation) unpack(sc, oh string) error {
	e := util.StringError("failed to unpack of %T", AllowedOperation{})
	if sc != "" {
		contract, err := NewAddressFromString(sc)
		if err != nil {
			return e.Wrap(err)
		}
		a.contract = contract
	}

	ht, err := hint.ParseHint(oh)
	if err != nil {
		return e.Wrap(err)
	}

	a.operation = ht
	return nil
}
