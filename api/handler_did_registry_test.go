package api

import (
	"net/http/httptest"
	"testing"
)

func TestDIDDocumentCacheKeyIncludesDIDQuery(t *testing.T) {
	path := "/did-registry/0xF5D0D9ef5e96676c57D5b6258f5849F72f05052bfca/document"
	reqA := httptest.NewRequest(
		"GET",
		path+"?did=did:imfact:0xF7206B8eFd2AbA1207E1578d5f3e05A2DCBeB3B5fca",
		nil,
	)
	reqB := httptest.NewRequest(
		"GET",
		path+"?did=did:imfact:0xead3598E10b583c6FFAEDC2eb64888d68Ed4d739fca",
		nil,
	)

	keyA := didDocumentCacheKey(reqA, ParseStringQuery(reqA.URL.Query().Get("did")))
	keyB := didDocumentCacheKey(reqB, ParseStringQuery(reqB.URL.Query().Get("did")))

	if keyA == keyB {
		t.Fatalf("expected different cache keys for different DID queries, got %q", keyA)
	}

	if keyA == CacheKeyPath(reqA) || keyB == CacheKeyPath(reqB) {
		t.Fatal("expected DID document cache key to include the did query")
	}
}
