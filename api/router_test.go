// Copyright (C) 2026 CRINE (https://www.crine.in) <support@crine.in>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"vyakhya/wordnet"
)

func TestAPIServer(t *testing.T) {
	idx := wordnet.NewIndex()
	// Load WordNet dataset
	err := idx.Load("../english-wordnet")
	if err != nil {
		t.Fatalf("Failed to load WordNet: %v", err)
	}

	server := NewServer(idx)
	handler := server.Handler()

	t.Run("Health Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse body: %v", err)
		}

		if resp["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %q", resp["status"])
		}
	})

	t.Run("Stats Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/stats", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var stats map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
			t.Fatalf("Failed to parse stats body: %v", err)
		}

		if _, exists := stats["total_words_indexed"]; !exists {
			t.Error("Expected key 'total_words_indexed' in stats response")
		}
		if _, exists := stats["uptime_seconds"]; !exists {
			t.Error("Expected key 'uptime_seconds' in stats response")
		}
	})

	t.Run("OpenAPI Spec Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/openapi.json", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var spec map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
			t.Fatalf("Failed to parse OpenAPI body: %v", err)
		}

		if spec["openapi"] != "3.0.3" {
			t.Errorf("Expected OpenAPI version 3.0.3, got %v", spec["openapi"])
		}
	})

	t.Run("Word Lookup 200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/word/happy", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check CORS headers
		origin := w.Header().Get("Access-Control-Allow-Origin")
		if origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin: *, got %q", origin)
		}

		var res wordnet.WordResult
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("Failed to parse lookup response: %v", err)
		}

		if res.Word != "happy" {
			t.Errorf("Expected word 'happy', got %q", res.Word)
		}
	})

	t.Run("Word Lookup 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/word/somefakeunrealnonexistentword", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}

		var resp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse 404 body: %v", err)
		}

		if resp["error"] != "word not found" {
			t.Errorf("Expected error message 'word not found', got %q", resp["error"])
		}
	})

	t.Run("CORS OPTIONS request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/v1/word/happy", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected OPTIONS status 204, got %d", w.Code)
		}
	})

	t.Run("Gzip Compression", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/word/happy", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentEncoding := w.Header().Get("Content-Encoding")
		if contentEncoding != "gzip" {
			t.Errorf("Expected Content-Encoding: gzip, got %q", contentEncoding)
		}
	})

	t.Run("Suggest Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/suggest?q=happ&limit=5", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var suggestions []string
		if err := json.Unmarshal(w.Body.Bytes(), &suggestions); err != nil {
			t.Fatalf("Failed to parse suggest body: %v", err)
		}

		if len(suggestions) == 0 {
			t.Error("Expected at least one suggestion, got none")
		}
	})

	t.Run("Random Endpoint Default", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/random", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var results []wordnet.WordResult
		if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
			t.Fatalf("Failed to parse random body: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 random word, got %d", len(results))
		}

		if results[0].Word == "" {
			t.Error("Expected word result to have a non-empty Word field")
		}
	})

	t.Run("Random Endpoint Alias", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/random", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var results []wordnet.WordResult
		if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
			t.Fatalf("Failed to parse random alias body: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 random word, got %d", len(results))
		}
	})

	t.Run("Random Endpoint with lim=5", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/random?lim=5", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var results []wordnet.WordResult
		if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
			t.Fatalf("Failed to parse random body: %v", err)
		}

		if len(results) != 5 {
			t.Errorf("Expected 5 random words, got %d", len(results))
		}

		// Ensure no duplicates in the response
		seen := make(map[string]bool)
		for _, res := range results {
			if res.Word == "" {
				t.Error("Expected word result to have a non-empty Word field")
			}
			if seen[res.Word] {
				t.Errorf("Found duplicate word %q in random results", res.Word)
			}
			seen[res.Word] = true
		}
	})

	t.Run("Random Endpoint cap at max 20", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/random?lim=50", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var results []wordnet.WordResult
		if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
			t.Fatalf("Failed to parse random body: %v", err)
		}

		if len(results) != 20 {
			t.Errorf("Expected capped 20 random words, got %d", len(results))
		}
	})

	t.Run("Landing Page Embedded Files", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}
