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
)

// Wikidata represents the wikidata field which can be either a single string or an array of strings in WordNet JSON.
type Wikidata []string

// UnmarshalJSON implements json.Unmarshaler to handle both string and list types.
func (w *Wikidata) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		var s []string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*w = s
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != "" {
		*w = []string{s}
	}
	return nil
}

// Example represents a usage example, which can be a simple text string or a structured object with source metadata.
type Example struct {
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler to handle both string and object types for examples.
func (e *Example) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '{' {
		type temp struct {
			Text   string `json:"text"`
			Source string `json:"source,omitempty"`
		}
		var t temp
		if err := json.Unmarshal(data, &t); err != nil {
			return err
		}
		e.Text = t.Text
		e.Source = t.Source
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	e.Text = s
	return nil
}

// Pronunciation represents phonetic values for words.
type Pronunciation struct {
	Value   string `json:"value"`
	Variety string `json:"variety,omitempty"`
}

// Sense represents sense-level metadata and relations inside word entries.
type Sense struct {
	ID          string   `json:"id"`
	Synset      string   `json:"synset"`
	AdjPosition string   `json:"adjposition,omitempty"`
	Agent       []string `json:"agent,omitempty"`
	Also        []string `json:"also,omitempty"`
	Antonym     []string `json:"antonym,omitempty"`
	BodyPart    []string `json:"body_part,omitempty"`
	ByMeansOf   []string `json:"by_means_of,omitempty"`
	Derivation  []string `json:"derivation,omitempty"`
	Destination []string `json:"destination,omitempty"`
	Event       []string `json:"event,omitempty"`
	Exemplifies []string `json:"exemplifies,omitempty"`
	Instrument  []string `json:"instrument,omitempty"`
	Location    []string `json:"location,omitempty"`
	Material    []string `json:"material,omitempty"`
	Participle  []string `json:"participle,omitempty"`
	Pertainym   []string `json:"pertainym,omitempty"`
	Property    []string `json:"property,omitempty"`
	Result      []string `json:"result,omitempty"`
	Sent        []string `json:"sent,omitempty"`
	Similar     []string `json:"similar,omitempty"`
	State       []string `json:"state,omitempty"`
	Subcat      []string `json:"subcat,omitempty"`
	Undergoer   []string `json:"undergoer,omitempty"`
	Uses        []string `json:"uses,omitempty"`
	Vehicle     []string `json:"vehicle,omitempty"`
}

// EntryInfo holds POS-specific lemmas/forms, pronunciations, and senses.
type EntryInfo struct {
	Form          []string        `json:"form,omitempty"`
	Pronunciation []Pronunciation `json:"pronunciation,omitempty"`
	Sense         []Sense         `json:"sense"`
}

// WordEntry maps parts of speech (e.g. "n", "v", "a", "r") to EntryInfo.
type WordEntry map[string]EntryInfo

// Synset represents the details of a synset (cognitive synonym group) from files like noun.animal.json.
type Synset struct {
	Definition    []string  `json:"definition"`
	PartOfSpeech  string    `json:"partOfSpeech"`
	ILI           string    `json:"ili,omitempty"`
	Members       []string  `json:"members"`
	Example       []Example `json:"example,omitempty"`
	Also          []string  `json:"also,omitempty"`
	Attribute     []string  `json:"attribute,omitempty"`
	Causes        []string  `json:"causes,omitempty"`
	DomainRegion  []string  `json:"domain_region,omitempty"`
	DomainTopic   []string  `json:"domain_topic,omitempty"`
	Entails       []string  `json:"entails,omitempty"`
	Exemplifies   []string  `json:"exemplifies,omitempty"`
	Hypernym      []string  `json:"hypernym,omitempty"`
	MeroMember    []string  `json:"mero_member,omitempty"`
	MeroPart      []string  `json:"mero_part,omitempty"`
	MeroSubstance []string  `json:"mero_substance,omitempty"`
	Similar       []string  `json:"similar,omitempty"`
	Source        string    `json:"source,omitempty"`
	Wikidata      Wikidata  `json:"wikidata,omitempty"`
}

// SimpleSynset contains basic information of a synset for inline relation resolution.
type SimpleSynset struct {
	ID         string   `json:"id"`
	Members    []string `json:"members"`
	Definition []string `json:"definition"`
}

// SimpleSense contains basic information of a sense for inline relation resolution.
type SimpleSense struct {
	ID         string   `json:"id"`
	Lemma      string   `json:"lemma"`
	SynsetID   string   `json:"synset_id"`
	Definition []string `json:"definition"`
}

// ResolvedRelation maps a relation name (e.g., "hypernym") to a list of resolved synsets or senses.
type ResolvedRelation struct {
	Relation string         `json:"relation"`
	Synsets  []SimpleSynset `json:"synsets,omitempty"`
	Senses   []SimpleSense  `json:"senses,omitempty"`
}

// ResolvedSense is a client-friendly representation of a sense, fully resolved.
type ResolvedSense struct {
	ID               string             `json:"id"`
	SynsetID         string             `json:"synset_id"`
	Definition       []string           `json:"definition"`
	Members          []string           `json:"members"` // Synonyms
	Examples         []Example          `json:"examples,omitempty"`
	ILI              string             `json:"ili,omitempty"`
	Wikidata         []string           `json:"wikidata,omitempty"`
	AdjPosition      string             `json:"adjposition,omitempty"`
	SentenceFrames   []string           `json:"sentence_frames,omitempty"` // Resolved from subcat via frames.json
	ExampleSentences []string           `json:"example_sentences,omitempty"` // From sent
	Relations        []ResolvedRelation `json:"relations,omitempty"`
}

// ResolvedEntry represents a part of speech and its associated senses.
type ResolvedEntry struct {
	PartOfSpeech   string          `json:"part_of_speech"`
	Forms          []string        `json:"forms,omitempty"`
	Pronunciations []Pronunciation `json:"pronunciations,omitempty"`
	Senses         []ResolvedSense `json:"senses"`
}

// WordResult is the final output returned by the dictionary API for a word query.
type WordResult struct {
	Word    string          `json:"word"`
	Entries []ResolvedEntry `json:"entries"`
}
