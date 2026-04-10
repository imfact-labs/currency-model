package digest_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/digest"
	digestisaac "github.com/imfact-labs/currency-model/digest/isaac"
	mongodbst "github.com/imfact-labs/currency-model/digest/mongodb"
	"github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestLoadManifestPreservesFeeAmounts(t *testing.T) {
	encs, benc := newTestEncoders(t)

	fee := []types.Amount{
		types.NewAmount(common.NewBig(10), types.CurrencyID("MCC")),
		types.NewAmount(common.NewBig(5), types.CurrencyID("USD")),
	}

	manifest := digestisaac.NewManifest(
		base.GenesisHeight,
		nil,
		valuehash.NewSHA256([]byte("proposal")),
		nil,
		nil,
		nil,
		time.Unix(123, 0).UTC(),
	)

	confirmedAt := time.Unix(456, 0).UTC()
	proposer := types.NewStringAddress("proposer")

	doc, err := digest.NewManifestDoc(
		manifest,
		benc,
		manifest.Height(),
		mongodbst.OperationItemInfo{TotalOperations: 1, ItemOperations: 1, Items: 1},
		fee,
		confirmedAt,
		proposer,
		base.Round(3),
		"test-build",
	)
	if err != nil {
		t.Fatalf("new manifest doc: %v", err)
	}

	b, err := doc.MarshalBSON()
	if err != nil {
		t.Fatalf("marshal manifest doc: %v", err)
	}

	gotManifest, operations, gotFee, gotConfirmedAt, gotProposer, gotRound, err := digest.LoadManifest(
		func(v interface{}) error {
			raw, ok := v.(*bson.Raw)
			if !ok {
				return fmt.Errorf("expected *bson.Raw, not %T", v)
			}

			*raw = bson.Raw(b)

			return nil
		},
		encs,
	)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	if gotManifest == nil {
		t.Fatal("expected manifest")
	}

	if gotManifest.Height() != manifest.Height() {
		t.Fatalf("unexpected manifest height: %v", gotManifest.Height())
	}

	if operations == nil || operations.TotalOperations != 1 || operations.ItemOperations != 1 || operations.Items != 1 {
		t.Fatalf("unexpected operations info: %+v", operations)
	}

	if len(gotFee) != len(fee) {
		t.Fatalf("unexpected fee length: %d", len(gotFee))
	}

	for i := range fee {
		if !gotFee[i].Equal(fee[i]) {
			t.Fatalf("unexpected fee[%d]: got=%v want=%v", i, gotFee[i], fee[i])
		}
	}

	if gotConfirmedAt != confirmedAt.String() {
		t.Fatalf("unexpected confirmed_at: %q", gotConfirmedAt)
	}

	if gotProposer != proposer.String() {
		t.Fatalf("unexpected proposer: %q", gotProposer)
	}

	if gotRound != 3 {
		t.Fatalf("unexpected round: %d", gotRound)
	}
}
