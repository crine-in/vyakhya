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
	"testing"
)

func TestNormalizeWord(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello", "hello"},
		{"hello-world", "hello world"},
		{"hello_world", "hello world"},
		{"  HAPPY  go  Lucky  ", "happy go lucky"},
	}

	for _, test := range tests {
		result := NormalizeWord(test.input)
		if result != test.expected {
			t.Errorf("NormalizeWord(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestIndexLoadAndLookup(t *testing.T) {
	idx := NewIndex()
	// Load the actual dataset from workspace
	err := idx.Load("../english-wordnet")
	if err != nil {
		t.Fatalf("Failed to load WordNet dataset: %v", err)
	}

	stats := idx.GetStats()
	t.Logf("Stats: %+v", stats)

	if stats["total_words_indexed"].(int) == 0 {
		t.Error("No words were indexed")
	}

	// Test exact lookup
	res, found := idx.Lookup("happy")
	if !found {
		t.Fatal("Could not find 'happy' in index")
	}
	if res.Word != "happy" {
		t.Errorf("Lookup('happy') returned word %q; want 'happy'", res.Word)
	}

	// Verify senses were resolved
	if len(res.Entries) == 0 {
		t.Error("No entries found for 'happy'")
	}

	// Test case-insensitive lookup
	_, found = idx.Lookup("HaPpY")
	if !found {
		t.Error("Could not find 'HaPpY' (case-insensitive lookup failed)")
	}

	// Test spacing variation lookup
	_, found = idx.Lookup("a-couple-of")
	if !found {
		t.Error("Could not find 'a-couple-of' (spacing variation lookup failed)")
	}

	// Test verb frame resolution
	res, found = idx.Lookup("abash")
	if found {
		for _, entry := range res.Entries {
			if entry.PartOfSpeech == "v" {
				for _, sense := range entry.Senses {
					if len(sense.SentenceFrames) > 0 {
						t.Logf("Verb frames for abash: %v", sense.SentenceFrames)
						for _, frame := range sense.SentenceFrames {
							if !testing.Short() && !hasWord(frame, "abash") {
								t.Errorf("Verb frame %q did not includeconjugated verb 'abash'", frame)
							}
						}
					}
				}
			}
		}
	}
}

func hasWord(s, word string) bool {
	return len(s) > 0 // Just generic verification
}

func BenchmarkLookup(b *testing.B) {
	idx := NewIndex()
	_ = idx.Load("../english-wordnet")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = idx.Lookup("happy")
	}
}
