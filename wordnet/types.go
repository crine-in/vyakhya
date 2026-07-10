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
	"sync"
)

var (
	internMu sync.Mutex
	interned = make(map[string]string)
)

// intern stores and retrieves a single copy of a string to reduce memory allocations.
func intern(s string) string {
	if s == "" {
		return ""
	}
	internMu.Lock()
	defer internMu.Unlock()
	if val, ok := interned[s]; ok {
		return val
	}
	interned[s] = s
	return s
}

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
	ID          string     `json:"id"`
	Synset      string     `json:"synset"`
	AdjPosition string     `json:"adjposition,omitempty"`
	Sent        []string   `json:"sent,omitempty"`
	Subcat      []string   `json:"subcat,omitempty"`
	Relations   []Relation `json:"-"`
}

// Relation represents a semantic relation mapping a relation name to its target IDs.
type Relation struct {
	Rel     string   `json:"rel"`
	Targets []string `json:"targets"`
}

// UnmarshalJSON implements json.Unmarshaler to parse Sense from a JSON object into our optimized representation.
func (s *Sense) UnmarshalJSON(data []byte) error {
	var raw struct {
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

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.ID = intern(raw.ID)
	s.Synset = intern(raw.Synset)
	s.AdjPosition = intern(raw.AdjPosition)
	
	if len(raw.Sent) > 0 {
		s.Sent = raw.Sent
		for i, sent := range s.Sent {
			s.Sent[i] = intern(sent)
		}
	}
	
	if len(raw.Subcat) > 0 {
		s.Subcat = raw.Subcat
		for i, sub := range s.Subcat {
			s.Subcat[i] = intern(sub)
		}
	}

	addRel := func(name string, targets []string) {
		if len(targets) > 0 {
			internedTargets := make([]string, len(targets))
			for i, t := range targets {
				internedTargets[i] = intern(t)
			}
			s.Relations = append(s.Relations, Relation{
				Rel:     intern(name),
				Targets: internedTargets,
			})
		}
	}

	addRel("agent", raw.Agent)
	addRel("also", raw.Also)
	addRel("antonym", raw.Antonym)
	addRel("body_part", raw.BodyPart)
	addRel("by_means_of", raw.ByMeansOf)
	addRel("derivation", raw.Derivation)
	addRel("destination", raw.Destination)
	addRel("event", raw.Event)
	addRel("exemplifies", raw.Exemplifies)
	addRel("instrument", raw.Instrument)
	addRel("location", raw.Location)
	addRel("material", raw.Material)
	addRel("participle", raw.Participle)
	addRel("pertainym", raw.Pertainym)
	addRel("property", raw.Property)
	addRel("result", raw.Result)
	addRel("similar", raw.Similar)
	addRel("state", raw.State)
	addRel("undergoer", raw.Undergoer)
	addRel("uses", raw.Uses)
	addRel("vehicle", raw.Vehicle)

	return nil
}

// EntryInfo holds POS-specific lemmas/forms, pronunciations, and senses.
type EntryInfo struct {
	Form          []string        `json:"form,omitempty"`
	Pronunciation []Pronunciation `json:"pronunciation,omitempty"`
	Sense         []Sense         `json:"sense"`
}

// POSEntry represents entry info for a specific part of speech.
type POSEntry struct {
	POS  string
	Info EntryInfo
}

// WordEntry represents all entries for a word across different parts of speech.
type WordEntry []POSEntry

// UnmarshalJSON implements json.Unmarshaler to handle decoding WordEntry from a JSON object.
func (we *WordEntry) UnmarshalJSON(data []byte) error {
	var raw map[string]EntryInfo
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*we = make([]POSEntry, 0, len(raw))
	for pos, info := range raw {
		for i, f := range info.Form {
			info.Form[i] = intern(f)
		}
		for i, p := range info.Pronunciation {
			info.Pronunciation[i].Value = intern(p.Value)
			info.Pronunciation[i].Variety = intern(p.Variety)
		}
		*we = append(*we, POSEntry{
			POS:  intern(pos),
			Info: info,
		})
	}
	return nil
}

// Synset represents the details of a synset (cognitive synonym group) from files like noun.animal.json.
type Synset struct {
	Definition   []string   `json:"definition"`
	PartOfSpeech string     `json:"partOfSpeech"`
	ILI          string     `json:"ili,omitempty"`
	Members      []string   `json:"members"`
	Example      []Example  `json:"example,omitempty"`
	Source       string     `json:"source,omitempty"`
	Wikidata     Wikidata   `json:"wikidata,omitempty"`
	Relations    []Relation `json:"-"`
}

// UnmarshalJSON implements json.Unmarshaler to parse Synset from a JSON object into our optimized representation.
func (s *Synset) UnmarshalJSON(data []byte) error {
	var raw struct {
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

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.Definition = raw.Definition
	for i, def := range s.Definition {
		s.Definition[i] = intern(def)
	}

	s.PartOfSpeech = intern(raw.PartOfSpeech)
	s.ILI = intern(raw.ILI)
	
	s.Members = raw.Members
	for i, m := range s.Members {
		s.Members[i] = intern(m)
	}

	s.Example = raw.Example
	for i, ex := range s.Example {
		s.Example[i].Text = intern(ex.Text)
		s.Example[i].Source = intern(ex.Source)
	}

	s.Source = intern(raw.Source)
	s.Wikidata = raw.Wikidata
	for i, w := range s.Wikidata {
		s.Wikidata[i] = intern(w)
	}

	addRel := func(name string, targets []string) {
		if len(targets) > 0 {
			internedTargets := make([]string, len(targets))
			for i, t := range targets {
				internedTargets[i] = intern(t)
			}
			s.Relations = append(s.Relations, Relation{
				Rel:     intern(name),
				Targets: internedTargets,
			})
		}
	}

	addRel("also", raw.Also)
	addRel("attribute", raw.Attribute)
	addRel("causes", raw.Causes)
	addRel("domain_region", raw.DomainRegion)
	addRel("domain_topic", raw.DomainTopic)
	addRel("entails", raw.Entails)
	addRel("exemplifies", raw.Exemplifies)
	addRel("hypernym", raw.Hypernym)
	addRel("mero_member", raw.MeroMember)
	addRel("mero_part", raw.MeroPart)
	addRel("mero_substance", raw.MeroSubstance)
	addRel("similar", raw.Similar)

	return nil
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
