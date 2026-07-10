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

package wordnet

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Index holds all parsed WordNet data and fast lookup indexes.
type Index struct {
	frames         map[string]string
	entries        map[string]WordEntry
	synsets        map[string]*Synset
	senses         map[string]senseRef
	normIndex      map[string][]string // Maps normalized word (lowercase, no hyphens/underscores) to original case words
	sortedNormKeys []string            // Sorted keys of normIndex for fast prefix search
	loadDuration   time.Duration
	mu             sync.RWMutex
}

type senseRef struct {
	Lemma    string
	SynsetID string
}

// NewIndex initializes an empty index.
func NewIndex() *Index {
	return &Index{
		frames:    make(map[string]string),
		entries:   make(map[string]WordEntry),
		synsets:   make(map[string]*Synset),
		senses:    make(map[string]senseRef),
		normIndex: make(map[string][]string),
	}
}

// NormalizeWord converts a string to lowercase and replaces underscores and hyphens with spaces.
func NormalizeWord(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)
	return strings.Join(words, " ")
}

// Load loads all WordNet data from the specified directory path.
func (idx *Index) Load(dirPath string) error {
	startTime := time.Now()

	// 1. Read frames.json
	framesPath := filepath.Join(dirPath, "frames.json")
	framesFile, err := os.Open(framesPath)
	if err != nil {
		return fmt.Errorf("failed to open frames.json: %w", err)
	}
	var rawFrames map[string]string
	if err := json.NewDecoder(framesFile).Decode(&rawFrames); err != nil {
		framesFile.Close()
		return fmt.Errorf("failed to parse frames.json: %w", err)
	}
	framesFile.Close()

	// Intern all keys and values in frames
	for k, v := range rawFrames {
		idx.frames[intern(k)] = intern(v)
	}

	// 2. Scan all JSON files in the directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read WordNet directory: %w", err)
	}

	// Separate entries files and synsets files
	var entryFiles []string
	var synsetFiles []string

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		name := file.Name()
		if name == "frames.json" {
			continue
		}

		if strings.HasPrefix(name, "entries-") {
			entryFiles = append(entryFiles, filepath.Join(dirPath, name))
		} else {
			synsetFiles = append(synsetFiles, filepath.Join(dirPath, name))
		}
	}

	// Load entries
	for _, path := range entryFiles {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open entry file %s: %w", path, err)
		}

		var fileEntries map[string]WordEntry
		if err := json.NewDecoder(file).Decode(&fileEntries); err != nil {
			file.Close()
			return fmt.Errorf("failed to parse entry file %s: %w", path, err)
		}
		file.Close()

		for word, wordEntry := range fileEntries {
			internedWord := intern(word)
			idx.entries[internedWord] = wordEntry

			// Normalize and index
			norm := intern(NormalizeWord(word))
			idx.normIndex[norm] = append(idx.normIndex[norm], internedWord)

			// Index senses for relation lookup
			for _, posInfo := range wordEntry {
				for _, sense := range posInfo.Info.Sense {
					if sense.ID != "" {
						idx.senses[sense.ID] = senseRef{
							Lemma:    internedWord,
							SynsetID: sense.Synset,
						}
					}
				}
			}
		}
	}

	// Load synsets
	for _, path := range synsetFiles {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open synset file %s: %w", path, err)
		}

		var fileSynsets map[string]*Synset
		if err := json.NewDecoder(file).Decode(&fileSynsets); err != nil {
			file.Close()
			return fmt.Errorf("failed to parse synset file %s: %w", path, err)
		}
		file.Close()

		for sid, synset := range fileSynsets {
			idx.synsets[intern(sid)] = synset
		}
	}

	// 3. Build and sort key list for autocomplete prefix searches
	idx.sortedNormKeys = make([]string, 0, len(idx.normIndex))
	for k := range idx.normIndex {
		idx.sortedNormKeys = append(idx.sortedNormKeys, k)
	}
	sort.Strings(idx.sortedNormKeys)

	idx.loadDuration = time.Since(startTime)

	// Explicit GC to reclaim parser overhead memory
	runtime.GC()

	return nil
}

// GetStats returns indexing stats.
func (idx *Index) GetStats() map[string]interface{} {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"total_words_indexed":   len(idx.entries),
		"total_synsets_indexed": len(idx.synsets),
		"total_senses_indexed":  len(idx.senses),
		"load_time_ms":          idx.loadDuration.Milliseconds(),
		"memory_allocated_mb":   float64(memStats.Alloc) / 1024 / 1024,
		"memory_system_mb":      float64(memStats.Sys) / 1024 / 1024,
	}
}

// Lookup queries a word in the index and returns a fully resolved result.
func (idx *Index) Lookup(word string) (*WordResult, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	norm := NormalizeWord(word)
	originalWords, exists := idx.normIndex[norm]
	if !exists {
		return nil, false
	}

	result := &WordResult{
		Word:    word,
		Entries: make([]ResolvedEntry, 0),
	}

	// Merge entries if query maps to multiple casing variants
	for _, origWord := range originalWords {
		wordEntry, exists := idx.entries[origWord]
		if !exists {
			continue
		}

		for _, posEntry := range wordEntry {
			pos := posEntry.POS
			entryInfo := posEntry.Info
			resolvedEntry := ResolvedEntry{
				PartOfSpeech:   pos,
				Forms:          entryInfo.Form,
				Pronunciations: entryInfo.Pronunciation,
				Senses:         make([]ResolvedSense, 0, len(entryInfo.Sense)),
			}

			for _, sense := range entryInfo.Sense {
				synset, found := idx.synsets[sense.Synset]
				if !found {
					continue
				}

				resSense := ResolvedSense{
					ID:             sense.ID,
					SynsetID:       sense.Synset,
					Definition:     synset.Definition,
					Members:        synset.Members,
					Examples:       synset.Example,
					ILI:            synset.ILI,
					Wikidata:       synset.Wikidata,
					AdjPosition:    sense.AdjPosition,
					Relations:      make([]ResolvedRelation, 0),
				}

				// Resolve verb frames
				if len(sense.Subcat) > 0 {
					resSense.SentenceFrames = make([]string, 0, len(sense.Subcat))
					for _, scat := range sense.Subcat {
						if template, ok := idx.frames[scat]; ok {
							// Format template with the current lemma
							formatted := strings.ReplaceAll(template, "----", origWord)
							resSense.SentenceFrames = append(resSense.SentenceFrames, formatted)
						}
					}
				}

				// Copy example sentences
				if len(sense.Sent) > 0 {
					resSense.ExampleSentences = sense.Sent
				}

				// 1. Resolve Synset-level relations
				for _, r := range synset.Relations {
					idx.resolveSynsetRelation(&resSense, r.Rel, r.Targets)
				}

				// 2. Resolve Sense-level relations
				for _, r := range sense.Relations {
					idx.resolveSenseRelation(&resSense, r.Rel, r.Targets)
				}

				resolvedEntry.Senses = append(resolvedEntry.Senses, resSense)
			}

			if len(resolvedEntry.Senses) > 0 {
				result.Entries = append(result.Entries, resolvedEntry)
			}
		}
	}

	if len(result.Entries) == 0 {
		return nil, false
	}

	return result, true
}

func (idx *Index) resolveSynsetRelation(resSense *ResolvedSense, relName string, targetIDs []string) {
	if len(targetIDs) == 0 {
		return
	}

	resolved := ResolvedRelation{
		Relation: relName,
		Synsets:  make([]SimpleSynset, 0, len(targetIDs)),
	}

	for _, tid := range targetIDs {
		if ts, ok := idx.synsets[tid]; ok {
			resolved.Synsets = append(resolved.Synsets, SimpleSynset{
				ID:         tid,
				Members:    ts.Members,
				Definition: ts.Definition,
			})
		}
	}

	if len(resolved.Synsets) > 0 {
		resSense.Relations = append(resSense.Relations, resolved)
	}
}

func (idx *Index) resolveSenseRelation(resSense *ResolvedSense, relName string, targetIDs []string) {
	if len(targetIDs) == 0 {
		return
	}

	resolved := ResolvedRelation{
		Relation: relName,
		Senses:   make([]SimpleSense, 0, len(targetIDs)),
	}

	for _, tid := range targetIDs {
		if sRef, ok := idx.senses[tid]; ok {
			var def []string
			if ts, ok := idx.synsets[sRef.SynsetID]; ok {
				def = ts.Definition
			}
			resolved.Senses = append(resolved.Senses, SimpleSense{
				ID:         tid,
				Lemma:      sRef.Lemma,
				SynsetID:   sRef.SynsetID,
				Definition: def,
			})
		}
	}

	if len(resolved.Senses) > 0 {
		resSense.Relations = append(resSense.Relations, resolved)
	}
}

// Suggest returns up to limit suggestions matching the given prefix.
func (idx *Index) Suggest(prefix string, limit int) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	norm := NormalizeWord(prefix)
	if norm == "" {
		return nil
	}

	// Binary search for the first key >= norm
	idxStart := sort.Search(len(idx.sortedNormKeys), func(i int) bool {
		return idx.sortedNormKeys[i] >= norm
	})

	results := make([]string, 0, limit)
	for i := idxStart; i < len(idx.sortedNormKeys); i++ {
		key := idx.sortedNormKeys[i]
		if !strings.HasPrefix(key, norm) {
			break // No more keys can have this prefix since slice is sorted
		}
		for _, orig := range idx.normIndex[key] {
			alreadyAdded := false
			for _, r := range results {
				if r == orig {
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				results = append(results, orig)
				if len(results) >= limit {
					return results
				}
			}
		}
	}
	return results
}

// Random returns a list of lim random WordResults from the index.
func (idx *Index) Random(lim int) []*WordResult {
	idx.mu.RLock()
	n := len(idx.sortedNormKeys)
	if n == 0 {
		idx.mu.RUnlock()
		return nil
	}

	if lim > n {
		lim = n
	}

	// Select random keys
	keys := make([]string, 0, lim)
	seen := make(map[string]bool)

	maxAttempts := lim * 5
	attempts := 0
	for len(keys) < lim && attempts < maxAttempts {
		attempts++
		randIdx := rand.IntN(n)
		key := idx.sortedNormKeys[randIdx]
		if !seen[key] {
			seen[key] = true
			keys = append(keys, key)
		}
	}
	idx.mu.RUnlock()

	// Now lookup the keys
	results := make([]*WordResult, 0, len(keys))
	for _, key := range keys {
		if res, ok := idx.Lookup(key); ok {
			results = append(results, res)
		}
	}
	return results
}
