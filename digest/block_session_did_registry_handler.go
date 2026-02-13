package digest

import (
	dstate "github.com/imfact-labs/imfact-currency/state/did-registry"
	"github.com/ProtoconNet/mitum2/base"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func PrepareDIDRegistry(bs *BlockSession, st base.State) (string, []mongo.WriteModel, error) {
	switch {
	case dstate.IsDesignStateKey(st.Key()):
		j, err := handleDIDRegistryDesignState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameDIDRegistry, j, nil
	case dstate.IsDataStateKey(st.Key()):
		j, err := handleDIDDataState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameDIDData, j, nil
	case dstate.IsDocumentStateKey(st.Key()):
		j, err := handleDIDDocumentState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameDIDDocument, j, nil
	}

	return "", nil, nil
}

func handleDIDRegistryDesignState(bs *BlockSession, st base.State) ([]mongo.WriteModel, error) {
	if DIDDesignDoc, err := NewDIDRegistryDesignDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(DIDDesignDoc),
		}, nil
	}
}

func handleDIDDataState(bs *BlockSession, st base.State) ([]mongo.WriteModel, error) {
	if DIDDataDoc, err := NewDIDDataDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(DIDDataDoc),
		}, nil
	}
}

func handleDIDDocumentState(bs *BlockSession, st base.State) ([]mongo.WriteModel, error) {
	if DIDDocumentDoc, err := NewDIDDocumentDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(DIDDocumentDoc),
		}, nil
	}
}
