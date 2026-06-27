package projectmemory

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"unicode"
)

// Retriever finds relevant memories for a free-text query, ordered most-relevant first.
// The Store is the default lexical implementation; HybridRetriever adds semantic ranking.
//
// SearchHit.Score is retriever-specific and meant for display/debugging, not comparison
// across retrievers: the lexical Store reports BM25 (lower is better) while the hybrid
// retriever reports a fused RRF score (higher is better). Always trust the result order.
type Retriever interface {
	Search(ctx context.Context, q SearchQuery) ([]SearchHit, error)
}

var _ Retriever = (*Store)(nil)

// trigramMinRunes is the shortest term a trigram index can match (three characters form
// one trigram); shorter terms are dropped from the substring stage.
const trigramMinRunes = 3

// Search ranks memories for a query using a lexical cascade: exact term match, then
// prefix match, then any-term (OR) prefix match on the unicode61 index, then substring
// match on the trigram index. Earlier, more precise stages rank first; later stages only
// fill remaining slots, and an id already returned is never repeated. The optional scope
// filter is parameter-driven so the SQL text never concatenates runtime values.
func (s *Store) Search(ctx context.Context, q SearchQuery) ([]SearchHit, error) {
	tokens := tokenizeQuery(q.Text)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("projectmemory: search requires a non-empty query: %w", ErrInvalidInput)
	}
	limit := q.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	scope := strings.TrimSpace(q.Scope)

	stages := []searchStage{
		{query: searchFTSQuery, expr: exactExpr(tokens)},
		{query: searchFTSQuery, expr: prefixExpr(tokens)},
		{query: searchFTSQuery, expr: orPrefixExpr(tokens)},
		{query: searchTrigramQuery, expr: trigramExpr(tokens)},
	}

	seen := make(map[string]struct{})
	ranExpr := make(map[string]struct{})
	results := make([]SearchHit, 0, limit)

	for _, stage := range stages {
		if len(results) >= limit {
			break
		}
		if stage.expr == "" {
			continue
		}
		dedupeKey := stage.query + "\x00" + stage.expr
		if _, done := ranExpr[dedupeKey]; done {
			continue
		}
		ranExpr[dedupeKey] = struct{}{}

		hits, err := s.runSearchStage(ctx, stage.query, stage.expr, scope, limit)
		if err != nil {
			return nil, err
		}
		for i := range hits {
			if _, ok := seen[hits[i].ID]; ok {
				continue
			}
			seen[hits[i].ID] = struct{}{}
			results = append(results, hits[i])
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

type searchStage struct {
	query string
	expr  string
}

// searchFTSQuery and searchTrigramQuery are identical except for the FTS table; both are
// static, with the optional scope filter driven by parameters.
const searchFTSQuery = `SELECT m.id, m.scope, m.mem_key, m.title, m.body, m.tags, m.source,
		m.created_at, m.updated_at, bm25(memories_fts) AS score
	FROM memories_fts
	JOIN memories m ON m.rowid = memories_fts.rowid
	WHERE memories_fts MATCH ?
	  AND (? = '' OR m.scope = ?)
	ORDER BY score ASC
	LIMIT ?`

const searchTrigramQuery = `SELECT m.id, m.scope, m.mem_key, m.title, m.body, m.tags, m.source,
		m.created_at, m.updated_at, bm25(memories_trigram) AS score
	FROM memories_trigram
	JOIN memories m ON m.rowid = memories_trigram.rowid
	WHERE memories_trigram MATCH ?
	  AND (? = '' OR m.scope = ?)
	ORDER BY score ASC
	LIMIT ?`

func (s *Store) runSearchStage(ctx context.Context, query, match, scope string, limit int) ([]SearchHit, error) {
	rows, err := s.db.QueryContext(ctx, query, match, scope, scope, limit)
	if err != nil {
		return nil, fmt.Errorf("projectmemory: search: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var hits []SearchHit
	for rows.Next() {
		var (
			memory  Memory
			key     sql.NullString
			tags    string
			created string
			updated string
			score   float64
		)
		if err := rows.Scan(
			&memory.ID, &memory.Scope, &key, &memory.Title, &memory.Body,
			&tags, &memory.Source, &created, &updated, &score,
		); err != nil {
			return nil, fmt.Errorf("projectmemory: scan search hit: %w", err)
		}
		hydrated, err := hydrateMemory(memory, key, tags, created, updated)
		if err != nil {
			return nil, err
		}
		hits = append(hits, SearchHit{Memory: hydrated, Score: score})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("projectmemory: iterate search: %w", err)
	}
	return hits, nil
}

// tokenizeQuery splits free-form text into alphanumeric tokens, dropping punctuation.
// Each token is safe to embed in a quoted FTS5 term, so the resulting MATCH expressions
// can never trigger an FTS5 syntax error from arbitrary user input.
func tokenizeQuery(text string) []string {
	raw := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	seen := make(map[string]struct{}, len(raw))
	tokens := make([]string, 0, len(raw))
	for _, token := range raw {
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}
	return tokens
}

// exactExpr requires every term (AND), each quoted to neutralize FTS5 operator keywords.
func exactExpr(tokens []string) string {
	return strings.Join(quoteTokens(tokens), " ")
}

// prefixExpr requires every term as a prefix (e.g. `"sqlit"*` matches "sqlite").
func prefixExpr(tokens []string) string {
	return strings.Join(prefixTokens(tokens), " ")
}

// orPrefixExpr matches any term as a prefix, broadening recall for loose paraphrases.
func orPrefixExpr(tokens []string) string {
	return strings.Join(prefixTokens(tokens), " OR ")
}

// trigramExpr matches any term as a substring via the trigram index, catching fragments
// inside identifiers (e.g. "sqlite" within "OpenSQLiteDatabase"). Terms shorter than a
// trigram are dropped because the index cannot match them.
func trigramExpr(tokens []string) string {
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len([]rune(token)) >= trigramMinRunes {
			parts = append(parts, `"`+token+`"`)
		}
	}
	return strings.Join(parts, " OR ")
}

func quoteTokens(tokens []string) []string {
	quoted := make([]string, len(tokens))
	for i, token := range tokens {
		quoted[i] = `"` + token + `"`
	}
	return quoted
}

func prefixTokens(tokens []string) []string {
	prefixed := make([]string, len(tokens))
	for i, token := range tokens {
		prefixed[i] = `"` + token + `"*`
	}
	return prefixed
}

const (
	// rrfK is the Reciprocal Rank Fusion constant. The standard default of 60 dampens the
	// weight of any single high rank, blending the two result lists smoothly.
	rrfK = 60
	// maxFusionCandidates bounds how many results each leg contributes to the fusion.
	maxFusionCandidates = 50
)

// HybridRetriever fuses the lexical cascade with semantic (embedding) ranking via
// Reciprocal Rank Fusion. RRF combines the two rankings without normalizing their
// incomparable scores (BM25 vs cosine). If embedding the query fails (e.g. Ollama is
// down), it degrades to lexical-only results so search never hard-fails.
type HybridRetriever struct {
	store    *Store
	embedder Embedder
}

var _ Retriever = (*HybridRetriever)(nil)

// NewHybridRetriever pairs a store with an embedder for semantic search.
func NewHybridRetriever(store *Store, embedder Embedder) *HybridRetriever {
	return &HybridRetriever{store: store, embedder: embedder}
}

// Search returns memories ranked by fused lexical and semantic relevance.
func (h *HybridRetriever) Search(ctx context.Context, q SearchQuery) ([]SearchHit, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	lexical, err := h.store.Search(ctx, SearchQuery{Text: q.Text, Scope: q.Scope, Limit: maxFusionCandidates})
	if err != nil {
		return nil, err
	}

	semantic, err := h.semanticSearch(ctx, q)
	if err != nil {
		// Graceful degradation: keep working on lexical results alone.
		slog.Default().Warn("project memory semantic search failed; using lexical only", "error", err)
		if len(lexical) > limit {
			return lexical[:limit], nil
		}
		return lexical, nil
	}

	return fuseRRF(lexical, semantic, limit), nil
}

// semanticSearch embeds the query and ranks every stored vector (for the active model) by
// cosine similarity. The curated corpus is small, so a brute-force scan is fast and avoids
// a native vector-index extension.
func (h *HybridRetriever) semanticSearch(ctx context.Context, q SearchQuery) ([]SearchHit, error) {
	if len(tokenizeQuery(q.Text)) == 0 {
		return nil, fmt.Errorf("projectmemory: search requires a non-empty query: %w", ErrInvalidInput)
	}

	queryVec, err := h.embedder.Embed(ctx, q.Text)
	if err != nil {
		return nil, err
	}

	candidates, err := h.store.vectorCandidates(ctx, strings.TrimSpace(q.Scope), h.embedder.Model())
	if err != nil {
		return nil, err
	}

	hits := make([]SearchHit, 0, len(candidates))
	for i := range candidates {
		hits = append(hits, SearchHit{
			Memory: candidates[i].Memory,
			Score:  cosineSimilarity(queryVec, candidates[i].vector),
		})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		return hits[i].Score > hits[j].Score
	})
	if len(hits) > maxFusionCandidates {
		hits = hits[:maxFusionCandidates]
	}
	return hits, nil
}

// fuseRRF merges two ranked lists by Reciprocal Rank Fusion and returns the top `limit`
// hits. The reported Score is the fused RRF weight (higher is better).
func fuseRRF(lexical, semantic []SearchHit, limit int) []SearchHit {
	scores := make(map[string]float64)
	memories := make(map[string]Memory)

	accumulate := func(list []SearchHit) {
		for rank := range list {
			id := list[rank].ID
			scores[id] += 1.0 / float64(rrfK+rank+1)
			if _, ok := memories[id]; !ok {
				memories[id] = list[rank].Memory
			}
		}
	}
	accumulate(lexical)
	accumulate(semantic)

	fused := make([]SearchHit, 0, len(scores))
	for id, score := range scores {
		fused = append(fused, SearchHit{Memory: memories[id], Score: score})
	}
	sort.SliceStable(fused, func(i, j int) bool {
		if fused[i].Score != fused[j].Score {
			return fused[i].Score > fused[j].Score
		}
		return fused[i].ID < fused[j].ID
	})
	if len(fused) > limit {
		fused = fused[:limit]
	}
	return fused
}
