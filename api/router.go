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
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vyakhya/wordnet"
)

// Server handles all API requests.
type Server struct {
	Index     *wordnet.Index
	StartTime time.Time
}

// NewServer creates a new API server.
func NewServer(idx *wordnet.Index) *Server {
	return &Server{
		Index:     idx,
		StartTime: time.Now(),
	}
}

// Handler returns the HTTP handler for the entire server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/stats", s.handleStats)
	mux.HandleFunc("GET /api/v1/word/{word}", s.handleWordLookup)
	mux.HandleFunc("GET /api/v1/suggest", s.handleSuggest)
	mux.HandleFunc("GET /api/v1/openapi.json", s.handleOpenAPI)

	// UI and docs endpoints
	mux.HandleFunc("GET /", s.handleLandingPage)

	// Apply middleware: Logging -> CORS -> Gzip
	return s.loggingMiddleware(s.corsMiddleware(s.gzipMiddleware(mux)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.Index.GetStats()
	stats["uptime_seconds"] = int(time.Since(s.StartTime).Seconds())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleWordLookup(w http.ResponseWriter, r *http.Request) {
	word := r.PathValue("word")
	word = strings.TrimSpace(word)
	if word == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"word parameter is required"}`))
		return
	}

	result, found := s.Index.Lookup(word)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"word not found"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	q = strings.TrimSpace(q)
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	suggestions := s.Index.Suggest(q, limit)
	if suggestions == nil {
		suggestions = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(suggestions)
}

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(OpenAPISpec))
}

// corsMiddleware injects CORS headers to allow cross-origin client usage.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter to compress responses.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

// gzipMiddleware compresses API and UI payloads using gzip compression.
func (s *Server) gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzw, r)
	})
}

// loggingMiddleware logs HTTP request details.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %s", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
	})
}

// OpenAPI 3.0 JSON Specification
const OpenAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Vyakhya Dictionary API",
    "description": "High performance, memory-efficient English Dictionary API powered by Open English WordNet.",
    "version": "1.0.0"
  },
  "paths": {
    "/health": {
      "get": {
        "summary": "Health Check",
        "description": "Get API service health status.",
        "responses": {
          "200": {
            "description": "Successful Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": { "type": "string", "example": "ok" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/stats": {
      "get": {
        "summary": "Service Statistics",
        "description": "Retrieve indexing, hardware, and performance statistics.",
        "responses": {
          "200": {
            "description": "Successful Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "total_words_indexed": { "type": "integer", "example": 128009 },
                    "total_synsets_indexed": { "type": "integer", "example": 107519 },
                    "total_senses_indexed": { "type": "integer", "example": 185129 },
                    "load_time_ms": { "type": "integer", "example": 1500 },
                    "memory_allocated_mb": { "type": "number", "example": 125.4 },
                    "memory_system_mb": { "type": "number", "example": 250.8 },
                    "uptime_seconds": { "type": "integer", "example": 3600 }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/suggest": {
      "get": {
        "summary": "Prefix Autocomplete Suggestions",
        "description": "Retrieve up to [limit] matching dictionary lemmas starting with a prefix query string.",
        "parameters": [
          {
            "name": "q",
            "in": "query",
            "required": true,
            "description": "Prefix query string to find matching suggestions",
            "schema": { "type": "string" }
          },
          {
            "name": "limit",
            "in": "query",
            "required": false,
            "description": "Max number of suggestions to return (default is 10)",
            "schema": { "type": "integer", "default": 10 }
          }
        ],
        "responses": {
          "200": {
            "description": "List of matching suggestions",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "type": "string" },
                  "example": ["happy", "happiness", "happily"]
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/word/{word}": {
      "get": {
        "summary": "Word Lookup",
        "description": "Query dictionary entries, definitions, sentence frames, examples, and semantic relations for a word.",
        "parameters": [
          {
            "name": "word",
            "in": "path",
            "required": true,
            "description": "The word to query (case-insensitive, handles spacing variations automatically)",
            "schema": { "type": "string" }
          }
        ],
        "responses": {
          "200": {
            "description": "Word found",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "word": { "type": "string", "example": "happy" },
                    "entries": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "properties": {
                          "part_of_speech": { "type": "string", "example": "a" },
                          "forms": {
                            "type": "array",
                            "items": { "type": "string" },
                            "example": ["happy"]
                          },
                          "pronunciations": {
                            "type": "array",
                            "items": {
                              "type": "object",
                              "properties": {
                                "value": { "type": "string", "example": "ˈhæpi" }
                              }
                            }
                          },
                          "senses": {
                            "type": "array",
                            "items": {
                              "type": "object",
                              "properties": {
                                "id": { "type": "string", "example": "happy%3:00:00::" },
                                "synset_id": { "type": "string", "example": "02194935-a" },
                                "definition": {
                                  "type": "array",
                                  "items": { "type": "string" },
                                  "example": ["enjoying or showing or marked by pleasure or satisfaction"]
                                },
                                "members": {
                                  "type": "array",
                                  "items": { "type": "string" },
                                  "example": ["happy"]
                                },
                                "examples": {
                                  "type": "array",
                                  "items": {
                                    "type": "object",
                                    "properties": {
                                      "text": { "type": "string", "example": "a happy smile" }
                                    }
                                  }
                                },
                                "relations": {
                                  "type": "array",
                                  "items": {
                                    "type": "object",
                                    "properties": {
                                      "relation": { "type": "string", "example": "antonym" },
                                      "senses": {
                                        "type": "array",
                                        "items": {
                                          "type": "object",
                                          "properties": {
                                            "id": { "type": "string", "example": "sad%3:00:00::" },
                                            "lemma": { "type": "string", "example": "sad" },
                                            "synset_id": { "type": "string", "example": "02193977-s" }
                                          }
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          },
          "400": {
            "description": "Bad Request",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": { "type": "string", "example": "word parameter is required" }
                  }
                }
              }
            }
          },
          "404": {
            "description": "Word not found",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "error": { "type": "string", "example": "word not found" }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`
