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
	framesBytes, err := os.ReadFile(framesPath)
	if err != nil {
		return fmt.Errorf("failed to read frames.json: %w", err)
	}
	if err := json.Unmarshal(framesBytes, &idx.frames); err != nil {
		return fmt.Errorf("failed to parse frames.json: %w", err)
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
		bytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read entry file %s: %w", path, err)
		}

		var fileEntries map[string]WordEntry
		if err := json.Unmarshal(bytes, &fileEntries); err != nil {
			return fmt.Errorf("failed to parse entry file %s: %w", path, err)
		}

		for word, wordEntry := range fileEntries {
			idx.entries[word] = wordEntry

			// Normalize and index
			norm := NormalizeWord(word)
			idx.normIndex[norm] = append(idx.normIndex[norm], word)

			// Index senses for relation lookup
			for _, posInfo := range wordEntry {
				for _, sense := range posInfo.Sense {
					if sense.ID != "" {
						idx.senses[sense.ID] = senseRef{
							Lemma:    word,
							SynsetID: sense.Synset,
						}
					}
				}
			}
		}
	}

	// Load synsets
	for _, path := range synsetFiles {
		bytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read synset file %s: %w", path, err)
		}

		var fileSynsets map[string]*Synset
		if err := json.Unmarshal(bytes, &fileSynsets); err != nil {
			return fmt.Errorf("failed to parse synset file %s: %w", path, err)
		}

		for sid, synset := range fileSynsets {
			idx.synsets[sid] = synset
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

		for pos, entryInfo := range wordEntry {
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
				idx.resolveSynsetRelation(&resSense, "also", synset.Also)
				idx.resolveSynsetRelation(&resSense, "attribute", synset.Attribute)
				idx.resolveSynsetRelation(&resSense, "causes", synset.Causes)
				idx.resolveSynsetRelation(&resSense, "domain_region", synset.DomainRegion)
				idx.resolveSynsetRelation(&resSense, "domain_topic", synset.DomainTopic)
				idx.resolveSynsetRelation(&resSense, "entails", synset.Entails)
				idx.resolveSynsetRelation(&resSense, "exemplifies", synset.Exemplifies)
				idx.resolveSynsetRelation(&resSense, "hypernym", synset.Hypernym)
				idx.resolveSynsetRelation(&resSense, "mero_member", synset.MeroMember)
				idx.resolveSynsetRelation(&resSense, "mero_part", synset.MeroPart)
				idx.resolveSynsetRelation(&resSense, "mero_substance", synset.MeroSubstance)
				idx.resolveSynsetRelation(&resSense, "similar", synset.Similar)

				// 2. Resolve Sense-level relations
				idx.resolveSenseRelation(&resSense, "agent", sense.Agent)
				idx.resolveSenseRelation(&resSense, "also", sense.Also)
				idx.resolveSenseRelation(&resSense, "antonym", sense.Antonym)
				idx.resolveSenseRelation(&resSense, "body_part", sense.BodyPart)
				idx.resolveSenseRelation(&resSense, "by_means_of", sense.ByMeansOf)
				idx.resolveSenseRelation(&resSense, "derivation", sense.Derivation)
				idx.resolveSenseRelation(&resSense, "destination", sense.Destination)
				idx.resolveSenseRelation(&resSense, "event", sense.Event)
				idx.resolveSenseRelation(&resSense, "exemplifies", sense.Exemplifies)
				idx.resolveSenseRelation(&resSense, "instrument", sense.Instrument)
				idx.resolveSenseRelation(&resSense, "location", sense.Location)
				idx.resolveSenseRelation(&resSense, "material", sense.Material)
				idx.resolveSenseRelation(&resSense, "participle", sense.Participle)
				idx.resolveSenseRelation(&resSense, "pertainym", sense.Pertainym)
				idx.resolveSenseRelation(&resSense, "property", sense.Property)
				idx.resolveSenseRelation(&resSense, "result", sense.Result)
				idx.resolveSenseRelation(&resSense, "similar", sense.Similar)
				idx.resolveSenseRelation(&resSense, "state", sense.State)
				idx.resolveSenseRelation(&resSense, "undergoer", sense.Undergoer)
				idx.resolveSenseRelation(&resSense, "uses", sense.Uses)
				idx.resolveSenseRelation(&resSense, "vehicle", sense.Vehicle)

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
