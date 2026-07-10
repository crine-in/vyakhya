# Vyakhya — High-Performance WordNet Dictionary API & Explorer

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg?style=flat&logo=go)](https://go.dev/)
[![Lookup Speed](https://img.shields.io/badge/Lookup_Speed-%3C4_microseconds-brightgreen.svg)]()
[![Memory Footprint](https://img.shields.io/badge/Memory_Footprint-~346_MB-orange.svg)]()

**Vyākhyā** (**व्याख्या**, **/ʋjɑː.kʰjɑː/**) is a dependency-free, high-performance, in-memory English Dictionary API and visual explorer built in Go, powered by the Open English WordNet JSON dataset.

It is designed to serve rich word definitions, synonyms, and multi-hop semantic relations with microsecond latency while keeping memory usage extremely optimized.

---

## ⚡ Performance Metrics

- **Average Lookup Latency**: **~3.2 microseconds** per query ($O(1)$ lookup complexity).
- **Throughput**: Handles **300,000+ requests per second** per CPU core.
- **Startup & Load Time**: Indexes the entire 73-file WordNet JSON dataset (~75MB raw JSON) in **~1.4 seconds**.
- **RAM Footprint**: **~346 MB RAM** in-memory post-garbage collection.

### 📈 HTTP API Stress Test Results
Under a local benchmark stress test (using Apache Bench `ab -n 100000 -c 100 -k` against the `/api/v1/word/happy` lookup endpoint):
- **Throughput**: **~12,963 requests per second** (mean).
- **Latency (mean)**: **0.077 ms** (across all concurrent requests).
- **99th Percentile Latency**: **10 ms** (99% of requests completed under 10ms).
- **Network Transfer Rate**: **~54.4 MB/sec**.

---

## ✨ Features

- **Zero-Dependency**: Built entirely on Go's standard library for maximum stability, portability, and compilation optimization.
- **Case and Symbol Normalization**: Automatically handles case variations, hyphens, and spaces (e.g., searching `HAPPY-GO-LUCKY`, `  happy   go   lucky  `, or `happy_go_lucky` resolves to the same entry).
- **Deep Semantic Relations**: Resolves and maps complex word-to-word and sense-to-sense relations (synonyms, antonyms, hypernyms, derivations, similar-to, causes, attributes, pertainyms, domain topics, and region classifications) in $O(1)$ time.
- **Verb Sentence Frames**: Automatically conjugates and conjugates verb templates to construct natural English usage frames (e.g., `"Somebody abashs somebody"`).
- **Polymorphic JSON Handling**: Tailored custom unmarshalers handle mixed WordNet schemas (like `wikidata` codes mapped to arrays/strings and `example` usage objects with/without source attribution).
- **Embedded UI Explorer**: A beautiful, dark-themed, glassmorphic search interface served at `/` with:
  - Live interactive word lookup.
  - Browser text-to-speech pronunciation.
  - Recursive navigation (clicking any synonym or relation badge searches for that word).
  - Real-time system health and memory usage metrics dashboard.
- **Embedded Swagger API Playground**: Access standard OpenAPI 3.0 docs and interactively query endpoints directly in the browser.
- **Production-Ready Middleware**: Pre-configured with native Gzip compression, CORS, JSON output headers, logging, and health checking.

---

## 🏗️ Architecture

```
vyakhya/
├── api/
│   ├── router.go          # HTTP routing, middleware (gzip, CORS, logging), and JSON openapi spec
│   ├── router_test.go     # HTTP endpoint unit tests and middleware verification
│   └── templates.go       # Embedded dark-themed glassmorphic lookup explorer & Swagger UI
├── english-wordnet/       # Directory containing WordNet JSON files
│   └── frames.json        # Verb subcategorization frame mappings
├── wordnet/
│   ├── index.go           # Index scanner, normalizer, and core lookup engine
│   ├── index_test.go      # Performance benchmarks and parser test suite
│   └── types.go           # Go structs and custom JSON unmarshalers
├── main.go                # Server entrypoint with flag configurations
├── LICENSE                # AGPL-3.0 License
├── Dockerfile             # Multi-stage containerized build configuration
└── README.md              # Project documentation
```

Vyakhya loads all JSON files concurrently at startup to construct an in-memory pointer map:

1.  **Word Normalization Map**: Maps normalized strings to exact entry indices.
2.  **Synset Index**: Pointers to synset contents for immediate dereferencing.
3.  **Sense Index**: Pointers to individual sense metadata for reverse relationships.

Once loaded, Vyakhya triggers a manual garbage collection call (`runtime.GC()`) to free parsing buffers, dropping idle RAM usage down to ~340MB.

---

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or later
- The Open English WordNet JSON dataset downloaded and placed in a directory (default folder: `./english-wordnet`).

### Clone and Run

1. Clone the repository to your local system:
   ```bash
   git clone https://github.com/crine-in/vyakhya.git
   cd vyakhya
   ```
2. Build and run the server:
   ```bash
   go run main.go -port 8080 -dataset ./english-wordnet
   ```
3. Open your browser and navigate to:
   - **Interactive UI Explorer**: `http://localhost:8080/`
   - **Swagger API Playground**: `http://localhost:8080/` (click the **API & Playground** tab)

---

## 🐳 Running with Docker

You can build and deploy Vyakhya as a lightweight, secure Docker container (packaged with alpine).

1. Build the Docker image:
   ```bash
   docker build -t vyakhya:latest .
   ```
2. Run the container:
   ```bash
   docker run -p 8080:8080 vyakhya:latest
   ```

---

## 🔌 API Endpoints

### 1. Word Lookup

`GET /api/v1/word/{word}`

Returns all matching entries, parts of speech, pronunciations, definitions, example sentences, verb frames, and semantic relation listings for a word.

#### Example Request

```bash
curl -H "Accept-Encoding: gzip" http://localhost:8080/api/v1/word/happy
```

#### Example Response

```json
{
  "word": "happy",
  "entries": [
    {
      "part_of_speech": "a",
      "forms": ["happy"],
      "pronunciations": [
        {
          "value": "ˈhæpi"
        }
      ],
      "senses": [
        {
          "id": "happy%3:00:00::",
          "synset_id": "02194935-a",
          "definition": [
            "enjoying or showing or marked by pleasure or satisfaction"
          ],
          "members": ["happy"],
          "examples": [
            {
              "text": "a happy smile"
            },
            {
              "text": "happy moods"
            }
          ],
          "relations": [
            {
              "relation": "antonym",
              "senses": [
                {
                  "id": "sad%3:00:00::",
                  "lemma": "sad",
                  "synset_id": "02193977-s"
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

### 2. System Metrics

`GET /api/v1/stats`

#### Example Request

```bash
curl http://localhost:8080/api/v1/stats
```

#### Example Response

```json
{
  "load_time_ms": 1390,
  "memory_allocated_mb": 346.002,
  "memory_system_mb": 456.45,
  "total_senses_indexed": 185129,
  "total_synsets_indexed": 107519,
  "total_words_indexed": 128009,
  "uptime_seconds": 120
}
```

---

## 🧪 Running Tests & Benchmarks

Run the complete test suite:

```bash
go test ./... -v
```

Run benchmarks to verify lookup performance:

```bash
go test -bench=. ./wordnet -benchmem
```

Run a local HTTP API stress test:
1. Compile and start the server in background silent mode (redirecting terminal output logs):
   ```bash
   go build -o vyakhya main.go
   ./vyakhya -port 8080 > /dev/null 2>&1 &
   ```
2. Run Apache Bench to simulate 100,000 requests with a concurrency of 100:
   ```bash
   ab -n 100000 -c 100 -k http://127.0.0.1:8080/api/v1/word/happy
   ```
3. Once completed, stop the background server:
   ```bash
   killall vyakhya
   ```

---

## 📜 License

Licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See the [LICENSE](LICENSE) file for details.

---

## 🏢 Contact & Company

Vyakhya is developed and maintained by **CRINE**.

- **Website**: [www.crine.in](https://www.crine.in)
- **General Inquiry**: [contact@crine.in](mailto:contact@crine.in)
- **Technical Support**: [support@crine.in](mailto:support@crine.in)
