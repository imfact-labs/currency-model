package digest

import (
	dstate "github.com/ProtoconNet/mitum-currency/v3/state/did-registry"
	"github.com/ProtoconNet/mitum2/base"
	"go.mongodb.org/mongo-driver/mongo"
)

func prepareDIDRegistry(bs *BlockSession, st base.State) (string, []mongo.WriteModel, error) {
	switch {
	case dstate.IsDesignStateKey(st.Key()):
		j, err := bs.handleDIDRegistryDesignState(st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameDIDRegistry, j, nil
	case dstate.IsDataStateKey(st.Key()):
		j, err := bs.handleDIDDataState(st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameDIDData, j, nil
	case dstate.IsDocumentStateKey(st.Key()):
		j, err := bs.handleDIDDocumentState(st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameDIDDocument, j, nil
	}

	return "", nil, nil
}

func (bs *BlockSession) handleDIDRegistryDesignState(st base.State) ([]mongo.WriteModel, error) {
	if DIDDesignDoc, err := NewDIDRegistryDesignDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(DIDDesignDoc),
		}, nil
	}
}

func (bs *BlockSession) handleDIDDataState(st base.State) ([]mongo.WriteModel, error) {
	if DIDDataDoc, err := NewDIDDataDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(DIDDataDoc),
		}, nil
	}
}

func (bs *BlockSession) handleDIDDocumentState(st base.State) ([]mongo.WriteModel, error) {
	if DIDDocumentDoc, err := NewDIDDocumentDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(DIDDocumentDoc),
		}, nil
	}
}
