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
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"vyakhya/api"
	"vyakhya/wordnet"
)

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	datasetPath := flag.String("dataset", "./english-wordnet", "Path to the english-wordnet JSON folder")
	flag.Parse()

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
