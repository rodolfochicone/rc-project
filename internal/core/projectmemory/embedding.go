package projectmemory

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// Embedder turns text into a dense vector for semantic retrieval. It is the optional
// second leg of the hybrid retriever; when absent, retrieval is purely lexical.
type Embedder interface {
	// Embed returns the embedding of text.
	Embed(ctx context.Context, text string) ([]float32, error)
	// Model identifies the producing model so stored vectors can be matched and a model
	// switch detected.
	Model() string
}

const (
	defaultOllamaEndpoint = "http://localhost:11434"
	defaultEmbedTimeout   = 30 * time.Second
)

// OllamaEmbedder calls a local Ollama daemon's /api/embed endpoint. It keeps data on the
// machine (no API key, no external network) at the cost of an external daemon.
type OllamaEmbedder struct {
	endpoint string
	model    string
	client   *http.Client
}

var _ Embedder = (*OllamaEmbedder)(nil)

// OllamaOption configures an OllamaEmbedder.
type OllamaOption func(*OllamaEmbedder)

// WithHTTPClient overrides the HTTP client (e.g. in tests).
func WithHTTPClient(client *http.Client) OllamaOption {
	return func(e *OllamaEmbedder) {
		if client != nil {
			e.client = client
		}
	}
}

// NewOllamaEmbedder builds an embedder for the given model. A blank endpoint defaults to
// the local Ollama address.
func NewOllamaEmbedder(endpoint, model string, opts ...OllamaOption) (*OllamaEmbedder, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("projectmemory: ollama embedder requires a model: %w", ErrInvalidInput)
	}
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		endpoint = defaultOllamaEndpoint
	}
	embedder := &OllamaEmbedder{
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: defaultEmbedTimeout},
	}
	for _, opt := range opts {
		opt(embedder)
	}
	return embedder, nil
}

// Model returns the configured model name.
func (e *OllamaEmbedder) Model() string {
	return e.model
}

// Embed requests a single embedding from Ollama.
func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	payload, err := json.Marshal(ollamaEmbedRequest{Model: e.model, Input: text})
	if err != nil {
		return nil, fmt.Errorf("projectmemory: marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint+"/api/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("projectmemory: build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("projectmemory: call ollama embed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("projectmemory: ollama embed returned status %d", resp.StatusCode)
	}

	var decoded ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("projectmemory: decode embed response: %w", err)
	}
	if len(decoded.Embeddings) == 0 || len(decoded.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("projectmemory: ollama returned no embedding for model %q", e.model)
	}
	return decoded.Embeddings[0], nil
}

type ollamaEmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// encodeVector serializes a vector as little-endian float32 for BLOB storage.
func encodeVector(vec []float32) []byte {
	buf := make([]byte, 4*len(vec))
	for i, value := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(value))
	}
	return buf
}

// decodeVector reverses encodeVector.
func decodeVector(blob []byte) ([]float32, error) {
	if len(blob)%4 != 0 {
		return nil, fmt.Errorf("projectmemory: corrupt vector blob of length %d", len(blob))
	}
	vec := make([]float32, len(blob)/4)
	for i := range vec {
		vec[i] = math.Float32frombits(binary.LittleEndian.Uint32(blob[i*4:]))
	}
	return vec, nil
}

// cosineSimilarity returns the cosine similarity of two equal-length vectors in [-1, 1];
// it returns 0 for mismatched lengths or a zero-magnitude vector.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
