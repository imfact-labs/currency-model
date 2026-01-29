package digest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ProtoconNet/mitum-currency/v3/digest/isaac"
	"github.com/ProtoconNet/mitum-currency/v3/digest/mongodb"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extras"
	ccstate "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	cestate "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/fixedtree"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var bulkWriteLimit = 500

type BlockSessionPrepareFunc func(*BlockSession, base.State) (string, []mongo.WriteModel, error)

type BlockSessioner interface {
	Prepare() error
	Commit(context.Context) error
	Close() error
}

type BlockSession struct {
	sync.RWMutex
	block           base.BlockMap
	ops             []base.Operation
	opsTree         fixedtree.Tree
	sts             []base.State
	st              *Database
	proposal        base.ProposalSignFact
	opsTreeNodes    map[string]base.OperationFixedtreeNode
	WriteModels     map[string][]mongo.WriteModel
	PrepareFunc     []BlockSessionPrepareFunc
	blockModels     []mongo.WriteModel
	operationModels []mongo.WriteModel
	statesValue     *sync.Map
	buildInfo       string
}

func NewBlockSession(
	st *Database, blk base.BlockMap, ops []base.Operation, opsTree fixedtree.Tree,
	sts []base.State, proposal base.ProposalSignFact, vs string) (
	*BlockSession, error,
) {
	if st.Readonly() {
		return nil, errors.Errorf("Readonly mode")
	}

	nst, err := st.New()
	if err != nil {
		return nil, err
	}

	return &BlockSession{
		st:          nst,
		block:       blk,
		ops:         ops,
		opsTree:     opsTree,
		sts:         sts,
		proposal:    proposal,
		WriteModels: make(map[string][]mongo.WriteModel),
		statesValue: &sync.Map{},

		buildInfo: vs,
	}, nil
}

func (bs *BlockSession) BlockMap() base.BlockMap {
	return bs.block
}

func (bs *BlockSession) Database() *Database {
	return bs.st
}

func (bs *BlockSession) Prepare() error {
	bs.Lock()
	defer bs.Unlock()

	if err := bs.prepareOperationsTree(); err != nil {
		return err
	}

	if err := bs.prepareBlock(); err != nil {
		return err
	}

	if err := bs.prepareOperations(); err != nil {
		return err
	}

	for i := range bs.sts {
		st := bs.sts[i]
		for _, prepareFunc := range bs.PrepareFunc {
			if colName, wrtModel, err := prepareFunc(bs, st); err != nil {
				return err
			} else if len(wrtModel) > 0 {
				_, ok := bs.WriteModels[colName]
				if !ok {
					bs.WriteModels[colName] = wrtModel
				} else {
					bs.WriteModels[colName] = append(bs.WriteModels[colName], wrtModel...)
				}
			}

		}
	}

	return nil
}

func (bs *BlockSession) Commit(ctx context.Context) error {
	bs.Lock()
	defer bs.Unlock()

	started := time.Now()
	defer func() {
		bs.statesValue.Store("commit", time.Since(started))
		_ = bs.close()
	}()

	_, err := bs.st.digestDB.Client().WithSession(func(txnCtx mongo.SessionContext, collection func(string) *mongo.Collection) (interface{}, error) {
		if err := bs.writeModels(txnCtx, DefaultColNameBlock, bs.blockModels); err != nil {
			return nil, err
		}

		if len(bs.operationModels) > 0 {
			if err := bs.writeModels(txnCtx, DefaultColNameOperation, bs.operationModels); err != nil {
				return nil, err
			}
		}

		for k, v := range bs.WriteModels {
			if len(v) > 0 {
				if err := bs.writeModels(txnCtx, k, v); err != nil {
					return nil, err
				}
			}
		}

		return nil, nil
	})

	return err
}

func (bs *BlockSession) Close() error {
	bs.Lock()
	defer bs.Unlock()

	return bs.close()
}

func (bs *BlockSession) prepareOperationsTree() error {
	nodes := map[string]base.OperationFixedtreeNode{}

	if err := bs.opsTree.Traverse(func(_ uint64, no fixedtree.Node) (bool, error) {
		nno := no.(base.OperationFixedtreeNode)
		if nno.Reason() == nil {
			nodes[nno.Key()] = nno
		} else {
			nodes[nno.Key()[:len(nno.Key())-1]] = nno
		}

		return true, nil
	}); err != nil {
		return err
	}

	bs.opsTreeNodes = nodes

	return nil
}

func (bs *BlockSession) prepareBlock() error {
	if bs.block == nil {
		return nil
	}

	var opInfo mongodbstorage.OperationItemInfo
	opInfo.TotalOperations = uint(len(bs.ops))

	var NoItemOperations uint
	var ItemOperations uint
	var Items uint
	for _, op := range bs.ops {
		feeable, ok := op.Fact().(extras.FeeAble)
		if ok {
			items, hasItem := feeable.FeeItemCount()
			if hasItem {
				ItemOperations++
				Items += uint(items)
			} else {
				NoItemOperations++
			}
		}
	}
	opInfo.NoItemOperations = NoItemOperations
	opInfo.ItemOperations = ItemOperations
	opInfo.Items = Items

	bs.blockModels = make([]mongo.WriteModel, 1)

	manifest := isaac.NewManifest(
		bs.block.Manifest().Height(),
		bs.block.Manifest().Previous(),
		bs.block.Manifest().Proposal(),
		bs.block.Manifest().OperationsTree(),
		bs.block.Manifest().StatesTree(),
		bs.block.Manifest().Suffrage(),
		bs.block.Manifest().ProposedAt(),
	)

	doc, err := NewManifestDoc(manifest, bs.st.digestDB.Encoder(), bs.block.Manifest().Height(), opInfo, bs.block.SignedAt(), bs.proposal.ProposalFact().Proposer(), bs.proposal.ProposalFact().Point().Round(), bs.buildInfo)
	if err != nil {
		return err
	}
	bs.blockModels[0] = mongo.NewInsertOneModel().SetDocument(doc)

	return nil
}

func (bs *BlockSession) prepareOperations() error {
	if len(bs.ops) < 1 {
		return nil
	}

	node := func(h util.Hash) (bool, bool, base.OperationProcessReasonError) {
		no, found := bs.opsTreeNodes[h.String()]
		if !found {
			return false, false, nil
		}

		return true, no.InState(), no.Reason()
	}

	bs.operationModels = make([]mongo.WriteModel, len(bs.ops))

	for i := range bs.ops {
		op := bs.ops[i]

		var doc OperationDoc
		switch found, inState, reason := node(op.Fact().Hash()); {
		case !found:
			return util.ErrNotFound.Errorf("Operation, %v in operations tree", op.Fact().Hash().String())
		default:
			var reasonMsg string
			switch {
			case reason == nil:
				reasonMsg = ""
			default:
				reasonMsg = reason.Msg()
			}
			d, err := NewOperationDoc(
				op,
				bs.st.digestDB.Encoder(),
				bs.block.Manifest().Height(),
				bs.block.SignedAt(),
				inState,
				reasonMsg,
				uint64(i),
			)
			if err != nil {
				return err
			}
			doc = d
		}

		bs.operationModels[i] = mongo.NewInsertOneModel().SetDocument(doc)
	}

	return nil
}

func PrepareAccounts(bs *BlockSession, st base.State) (string, []mongo.WriteModel, error) {
	switch {
	case ccstate.IsAccountStateKey(st.Key()):
		j, err := handleAccountState(bs, st)
		if err != nil {
			return "", nil, err
		}
		return DefaultColNameAccount, j, nil
	case ccstate.IsBalanceStateKey(st.Key()):
		j, _, err := handleBalanceState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameBalance, j, nil
	case cestate.IsStateContractAccountKey(st.Key()):
		j, err := handleContractAccountState(bs, st)
		if err != nil {
			return "", nil, err
		}
		return DefaultColNameContractAccount, j, nil
	}

	return "", nil, nil
}

func PrepareCurrencies(bs *BlockSession, st base.State) (string, []mongo.WriteModel, error) {
	switch {
	case ccstate.IsDesignStateKey(st.Key()):
		j, err := handleCurrencyState(bs, st)
		if err != nil {
			return "", nil, err
		}

		return DefaultColNameCurrency, j, nil
	}

	return "", nil, nil
}

func handleAccountState(bs *BlockSession, st base.State) ([]mongo.WriteModel, error) {
	if rs, err := NewAccountValue(st); err != nil {
		return nil, err
	} else if doc, err := NewAccountDoc(rs, bs.st.digestDB.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
	}
}

func handleBalanceState(bs *BlockSession, st base.State) ([]mongo.WriteModel, string, error) {
	doc, address, err := NewBalanceDoc(st, bs.st.digestDB.Encoder())
	if err != nil {
		return nil, "", err
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, address, nil
}

func handleContractAccountState(bs *BlockSession, st base.State) ([]mongo.WriteModel, error) {
	doc, err := NewContractAccountStatusDoc(st, bs.st.digestDB.Encoder())
	if err != nil {
		return nil, err
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func handleCurrencyState(bs *BlockSession, st base.State) ([]mongo.WriteModel, error) {
	doc, err := NewCurrencyDoc(st, bs.st.digestDB.Encoder())
	if err != nil {
		return nil, err
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func (bs *BlockSession) writeModels(ctx context.Context, col string, models []mongo.WriteModel) error {
	started := time.Now()
	defer func() {
		bs.statesValue.Store(fmt.Sprintf("write-models-%s", col), time.Since(started))
	}()

	n := len(models)
	if n < 1 {
		return nil
	} else if n <= bulkWriteLimit {
		return bs.writeModelsChunk(ctx, col, models)
	}

	z := n / bulkWriteLimit
	if n%bulkWriteLimit != 0 {
		z++
	}

	for i := 0; i < z; i++ {
		s := i * bulkWriteLimit
		e := s + bulkWriteLimit
		if e > n {
			e = n
		}

		if err := bs.writeModelsChunk(ctx, col, models[s:e]); err != nil {
			return err
		}
	}

	return nil
}

func (bs *BlockSession) writeModelsChunk(ctx context.Context, col string, models []mongo.WriteModel) error {
	opts := options.BulkWrite().SetOrdered(false)
	if res, err := bs.st.digestDB.Client().Collection(col).BulkWrite(ctx, models, opts); err != nil {
		return err
	} else if res != nil && res.InsertedCount < 1 {
		return errors.Errorf("Not inserted to %s", col)
	}

	return nil
}

func (bs *BlockSession) close() error {
	bs.block = nil
	bs.ops = nil
	bs.opsTree = fixedtree.EmptyTree()
	bs.WriteModels = nil
	bs.sts = nil
	bs.proposal = nil
	bs.opsTreeNodes = nil
	bs.blockModels = nil
	bs.operationModels = nil

	return bs.st.Close()
}
