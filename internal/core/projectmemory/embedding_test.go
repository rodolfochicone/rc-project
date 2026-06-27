package projectmemory

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestEncodeDecodeVectorRoundTrip(t *testing.T) {
	t.Parallel()
	in := []float32{0, 1, -0.5, 3.14159, 1e9, -42}

	out, err := decodeVector(encodeVector(in))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch: in=%v out=%v", in, out)
	}
}

func TestDecodeVectorRejectsCorruptLength(t *testing.T) {
	t.Parallel()
	if _, err := decodeVector([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected error for non-multiple-of-4 blob length")
	}
}

func TestCosineSimilarity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		a, b []float32
		want float64
	}{
		{"identical", []float32{1, 2, 3}, []float32{1, 2, 3}, 1},
		{"orthogonal", []float32{1, 0}, []float32{0, 1}, 0},
		{"opposite", []float32{1, 0}, []float32{-1, 0}, -1},
		{"mismatched length", []float32{1, 0}, []float32{1}, 0},
		{"zero magnitude", []float32{0, 0}, []float32{1, 1}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := cosineSimilarity(tc.a, tc.b); math.Abs(got-tc.want) > 1e-6 {
				t.Fatalf("cosineSimilarity = %f, want %f", got, tc.want)
			}
		})
	}
}

func TestOllamaEmbedderEmbeds(t *testing.T) {
	t.Parallel()

	var gotModel, gotInput string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var req ollamaEmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		gotModel, gotInput = req.Model, req.Input
		_ = json.NewEncoder(w).Encode(ollamaEmbedResponse{Embeddings: [][]float32{{0.1, 0.2, 0.3}}})
	}))
	t.Cleanup(srv.Close)

	embedder, err := NewOllamaEmbedder(srv.URL, "embeddinggemma")
	if err != nil {
		t.Fatalf("new embedder: %v", err)
	}

	vec, err := embedder.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("embed: %v", err)
	}
	if !reflect.DeepEqual(vec, []float32{0.1, 0.2, 0.3}) {
		t.Fatalf("vector = %v, want [0.1 0.2 0.3]", vec)
	}
	if gotModel != "embeddinggemma" || gotInput != "hello world" {
		t.Fatalf("request carried model=%q input=%q", gotModel, gotInput)
	}
	if embedder.Model() != "embeddinggemma" {
		t.Fatalf("Model() = %q", embedder.Model())
	}
}

func TestOllamaEmbedderErrorsOnBadStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	embedder, err := NewOllamaEmbedder(srv.URL, "m")
	if err != nil {
		t.Fatalf("new embedder: %v", err)
	}
	if _, err := embedder.Embed(context.Background(), "x"); err == nil {
		t.Fatal("expected error on 500 status")
	}
}

func TestNewOllamaEmbedderRequiresModel(t *testing.T) {
	t.Parallel()
	if _, err := NewOllamaEmbedder("http://localhost:11434", "  "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}
