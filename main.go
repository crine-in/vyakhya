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

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"vyakhya/api"
	"vyakhya/wordnet"
)

// loadEnv parses a .env file and sets environment variables if not already set.
func loadEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return // Ignore if file doesn't exist
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip quotes if present
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func main() {
	loadEnv(".env")

	defaultPort := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		defaultPort = envPort
	}

	port := flag.String("port", defaultPort, "Port to run the server on")
	datasetPath := flag.String("dataset", "./english-wordnet", "Path to the english-wordnet JSON folder")
	flag.Parse()

	// Apply Memory Limits from Env
	if memLimitStr := os.Getenv("MEM_LIMIT_MB"); memLimitStr != "" {
		if limitMB, err := strconv.Atoi(memLimitStr); err == nil && limitMB > 0 {
			limitBytes := int64(limitMB) * 1024 * 1024
			debug.SetMemoryLimit(limitBytes)
			log.Printf("Security: Memory limit set to %d MB (%d bytes)", limitMB, limitBytes)
		}
	}

	// Apply GC Target from Env
	if gcPercentStr := os.Getenv("GC_PERCENT"); gcPercentStr != "" {
		if percent, err := strconv.Atoi(gcPercentStr); err == nil && percent > 0 {
			debug.SetGCPercent(percent)
			log.Printf("Security: Go GC target percent set to %d%%", percent)
		}
	}

	log.Println("Initializing Vyakhya Dictionary Indexer...")
	idx := wordnet.NewIndex()

	log.Printf("Loading WordNet dataset from: %s", *datasetPath)
	startTime := time.Now()
	if err := idx.Load(*datasetPath); err != nil {
		log.Fatalf("Fatal: Failed to load dataset: %v", err)
	}

	stats := idx.GetStats()
	log.Printf("Indexer loaded successfully in %v!", time.Since(startTime))
	log.Printf("Indexed Words:  %s", formatInt(stats["total_words_indexed"].(int)))
	log.Printf("Indexed Synsets: %s", formatInt(stats["total_synsets_indexed"].(int)))
	log.Printf("Indexed Senses:  %s", formatInt(stats["total_senses_indexed"].(int)))
	log.Printf("Memory Usage:    %.2f MB (sys: %.2f MB)", stats["memory_allocated_mb"].(float64), stats["memory_system_mb"].(float64))

	server := api.NewServer(idx)
	handler := server.Handler()

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Vyakhya Dictionary Service is listening on http://localhost:%s", *port)
	log.Printf("Interactive lookup interface available at http://localhost:%s/", *port)
	log.Printf("API Swagger Playground available at http://localhost:%s/ (Docs tab)", *port)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Fatal: Server failed to start: %v", err)
	}
}

// formatInt formats integers with comma separators.
func formatInt(n int) string {
	in := fmt.Sprintf("%d", n)
	out := make([]byte, len(in)+(len(in)-2)/3)
	if len(in) <= 3 {
		return in
	}
	j := 0
	for i := 0; i < len(in); i++ {
		if i > 0 && (len(in)-i)%3 == 0 {
			out[j] = ','
			j++
		}
		out[j] = in[i]
		j++
	}
	return string(out)
}
