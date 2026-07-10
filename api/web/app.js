// State management
let bookmarks = [];
let searchHistory = [];
let activeSuggestionIndex = -1;
let currentSuggestions = [];
let currentWordData = null;
let swaggerInitialized = false;

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    loadBookmarks();
    loadHistory();
    fetchStats(); // Fetch statistics once at page load

    // Setup clear search button logic
    const searchInput = document.getElementById('search-input');
    const clearBtn = document.getElementById('clear-btn');
    
    searchInput.addEventListener('input', () => {
        const query = searchInput.value.trim();
        clearBtn.style.display = query.length > 0 ? 'block' : 'none';
        debouncedGetSuggestions(query);
    });

    // Close suggestions dropdown when clicking outside
    document.addEventListener('click', (e) => {
        if (e.target.id !== 'search-input' && !e.target.closest('.suggestions-dropdown')) {
            hideSuggestions();
        }
    });
});

// Tab navigation
function switchTab(tabId) {
    document.querySelectorAll('.nav-btn').forEach(btn => btn.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));

    if (tabId === 'explorer') {
        document.getElementById('btn-explorer').classList.add('active');
        document.getElementById('tab-explorer').classList.add('active');
    } else if (tabId === 'docs') {
        document.getElementById('btn-docs').classList.add('active');
        document.getElementById('tab-docs').classList.add('active');
        initSwagger();
    }
}

// Local Storage for Bookmarks
function loadBookmarks() {
    try {
        const saved = localStorage.getItem('vyakhya_bookmarks');
        bookmarks = saved ? JSON.parse(saved) : [];
    } catch (e) {
        bookmarks = [];
    }
    renderBookmarks();
}

function saveBookmarks() {
    localStorage.setItem('vyakhya_bookmarks', JSON.stringify(bookmarks));
    renderBookmarks();
    updateStarUI();
}

// Toggle a word bookmark
function toggleBookmark(word) {
    if (!word) return;
    const w = word.toLowerCase();
    const idx = bookmarks.indexOf(w);
    if (idx === -1) {
        bookmarks.push(w);
    } else {
        bookmarks.splice(idx, 1);
    }
    saveBookmarks();
}

function renderBookmarks() {
    const list = document.getElementById('bookmarks-list');
    const countBadge = document.getElementById('bookmarks-count');
    
    countBadge.innerText = bookmarks.length;
    
    if (bookmarks.length === 0) {
        list.innerHTML = '<li class="empty-state">No bookmarked words</li>';
        return;
    }

    list.innerHTML = '';
    // Display bookmarks sorted alphabetically
    [...bookmarks].sort().forEach(word => {
        const li = document.createElement('li');
        const btn = document.createElement('button');
        btn.className = 'list-item-btn';
        btn.innerText = word;
        btn.onclick = () => performSearch(word);
        li.appendChild(btn);
        list.appendChild(li);
    });
}

function updateStarUI() {
    const starBtn = document.getElementById('bookmark-star');
    if (!starBtn || !currentWordData) return;
    
    const word = currentWordData.word.toLowerCase();
    if (bookmarks.includes(word)) {
        starBtn.classList.add('starred');
        starBtn.innerHTML = '&#9733;'; // Filled star
        starBtn.title = 'Remove from Bookmarks';
    } else {
        starBtn.classList.remove('starred');
        starBtn.innerHTML = '&#9734;'; // Empty star
        starBtn.title = 'Add to Bookmarks';
    }
}

// Local Storage for Search History
function loadHistory() {
    try {
        const saved = localStorage.getItem('vyakhya_history');
        searchHistory = saved ? JSON.parse(saved) : [];
    } catch (e) {
        searchHistory = [];
    }
    renderHistory();
}

function saveHistory() {
    localStorage.setItem('vyakhya_history', JSON.stringify(searchHistory));
    renderHistory();
}

function addToHistory(word) {
    if (!word) return;
    const w = word.toLowerCase();
    
    // Remove existing item to put it at the top
    searchHistory = searchHistory.filter(item => item !== w);
    searchHistory.unshift(w);
    
    // Limit history length to 10 items
    if (searchHistory.length > 10) {
        searchHistory.pop();
    }
    
    saveHistory();
}

function clearHistory() {
    searchHistory = [];
    saveHistory();
}

function renderHistory() {
    const list = document.getElementById('history-list');
    
    if (searchHistory.length === 0) {
        list.innerHTML = '<li class="empty-state">No recent searches</li>';
        return;
    }

    list.innerHTML = '';
    searchHistory.forEach(word => {
        const li = document.createElement('li');
        const btn = document.createElement('button');
        btn.className = 'list-item-btn';
        btn.innerText = word;
        btn.onclick = () => performSearch(word);
        li.appendChild(btn);
        list.appendChild(li);
    });
}

// Telemetry Stats (No polling - fetch on load or manual refresh)
let statsLoading = false;
async function fetchStats() {
    if (statsLoading) return;
    statsLoading = true;
    
    const refreshBtn = document.querySelector('.refresh-btn');
    if (refreshBtn) refreshBtn.innerText = 'LOADING';

    try {
        const response = await fetch('/api/v1/stats');
        if (response.ok) {
            const data = await response.json();
            document.getElementById('stat-words').innerText = data.total_words_indexed.toLocaleString();
            document.getElementById('stat-synsets').innerText = data.total_synsets_indexed.toLocaleString();
            document.getElementById('stat-senses').innerText = data.total_senses_indexed.toLocaleString();
            document.getElementById('stat-load').innerText = data.load_time_ms + ' ms';
            document.getElementById('stat-alloc').innerText = data.memory_allocated_mb.toFixed(2) + ' MB';
            document.getElementById('stat-sys').innerText = data.memory_system_mb.toFixed(2) + ' MB';
            
            const hours = Math.floor(data.uptime_seconds / 3600);
            const minutes = Math.floor((data.uptime_seconds % 3600) / 60);
            const seconds = data.uptime_seconds % 60;
            document.getElementById('stat-uptime').innerText = 
                (hours > 0 ? hours + 'h ' : '') + minutes + 'm ' + seconds + 's';
        }
    } catch (err) {
        console.error("Failed to query stats:", err);
    } finally {
        statsLoading = false;
        if (refreshBtn) refreshBtn.innerText = 'REFRESH';
    }
}

// Autocomplete Suggestions logic
let suggestionsTimeout = null;
function debouncedGetSuggestions(query) {
    if (suggestionsTimeout) clearTimeout(suggestionsTimeout);
    
    if (query.length < 1) {
        hideSuggestions();
        return;
    }

    suggestionsTimeout = setTimeout(async () => {
        try {
            const response = await fetch(`/api/v1/suggest?q=${encodeURIComponent(query)}&limit=10`);
            if (response.ok) {
                const data = await response.json();
                currentSuggestions = data;
                showSuggestions(data);
            }
        } catch (e) {
            console.error("Autocomplete fetch error", e);
        }
    }, 100);
}

function showSuggestions(list) {
    const dropdown = document.getElementById('suggestions-box');
    if (list.length === 0) {
        hideSuggestions();
        return;
    }

    dropdown.innerHTML = '';
    activeSuggestionIndex = -1;

    list.forEach((word, index) => {
        const item = document.createElement('div');
        item.className = 'suggestion-item';
        item.innerText = word;
        item.onclick = () => {
            document.getElementById('search-input').value = word;
            performSearch(word);
            hideSuggestions();
        };
        dropdown.appendChild(item);
    });

    dropdown.style.display = 'flex';
}

function hideSuggestions() {
    const dropdown = document.getElementById('suggestions-box');
    dropdown.style.display = 'none';
    activeSuggestionIndex = -1;
    currentSuggestions = [];
}

// Handles Up/Down arrows and Enter navigation inside suggestions dropdown
function handleSearchInputKey(event) {
    const dropdown = document.getElementById('suggestions-box');
    const items = dropdown.getElementsByClassName('suggestion-item');
    
    if (event.key === 'ArrowDown') {
        event.preventDefault();
        if (items.length === 0) return;
        
        if (activeSuggestionIndex >= 0) {
            items[activeSuggestionIndex].classList.remove('active');
        }
        
        activeSuggestionIndex = (activeSuggestionIndex + 1) % items.length;
        items[activeSuggestionIndex].classList.add('active');
        document.getElementById('search-input').value = currentSuggestions[activeSuggestionIndex];
    } else if (event.key === 'ArrowUp') {
        event.preventDefault();
        if (items.length === 0) return;
        
        if (activeSuggestionIndex >= 0) {
            items[activeSuggestionIndex].classList.remove('active');
        }
        
        if (activeSuggestionIndex <= 0) {
            activeSuggestionIndex = items.length - 1;
        } else {
            activeSuggestionIndex--;
        }
        items[activeSuggestionIndex].classList.add('active');
        document.getElementById('search-input').value = currentSuggestions[activeSuggestionIndex];
    } else if (event.key === 'Enter') {
        event.preventDefault();
        const query = document.getElementById('search-input').value.trim();
        if (query) {
            performSearch(query);
            hideSuggestions();
        }
    } else if (event.key === 'Escape') {
        hideSuggestions();
    }
}

function clearSearchInput() {
    const input = document.getElementById('search-input');
    input.value = '';
    input.focus();
    document.getElementById('clear-btn').style.display = 'none';
    hideSuggestions();
}

// Search and API Query execution
async function performSearch(word) {
    const searchInput = document.getElementById('search-input');
    const query = word || searchInput.value.trim();
    if (!query) return;

    searchInput.value = query;
    document.getElementById('clear-btn').style.display = 'block';
    hideSuggestions();

    const resultsArea = document.getElementById('results-area');
    resultsArea.innerHTML = '<div class="welcome-screen"><div class="welcome-heading">Searching Database...</div><div class="welcome-desc">Querying WordNet lexical trees...</div></div>';

    try {
        const response = await fetch(`/api/v1/word/${encodeURIComponent(query)}`);
        if (!response.ok) {
            if (response.status === 404) {
                resultsArea.innerHTML = `
                    <div class="welcome-screen">
                        <div class="welcome-heading" style="color: #666;">Word Not Found</div>
                        <div class="welcome-desc">“${escapeHTML(query)}” is not registered in the WordNet lexicon. Double check spelling variations.</div>
                    </div>`;
            } else {
                resultsArea.innerHTML = `
                    <div class="welcome-screen">
                        <div class="welcome-heading" style="color: #aa4444;">Error</div>
                        <div class="welcome-desc">A server-side error occurred while resolving this lexical index.</div>
                    </div>`;
            }
            return;
        }

        const data = await response.json();
        currentWordData = data;
        renderResult(data);
        addToHistory(data.word);
    } catch (err) {
        console.error("Search error:", err);
        resultsArea.innerHTML = `
            <div class="welcome-screen">
                <div class="welcome-heading" style="color: #aa4444;">Connection Failed</div>
                <div class="welcome-desc">Unable to establish connection to the Vyākhyā API server. Verify server status.</div>
            </div>`;
    }
}

// Render word lookups inside results area
function renderResult(data) {
    const resultsArea = document.getElementById('results-area');
    resultsArea.innerHTML = '';

    // Word Metadata Header
    const header = document.createElement('div');
    header.className = 'word-meta-header';

    const mainInfo = document.createElement('div');
    mainInfo.className = 'word-main-info';

    const headingRow = document.createElement('div');
    headingRow.className = 'word-heading-row';

    const h1 = document.createElement('h1');
    h1.innerText = data.word.toLowerCase();
    headingRow.appendChild(h1);

    const starBtn = document.createElement('button');
    starBtn.id = 'bookmark-star';
    starBtn.className = 'star-btn';
    starBtn.onclick = () => toggleBookmark(data.word);
    headingRow.appendChild(starBtn);
    mainInfo.appendChild(headingRow);

    // Phonetic representation
    let pronVal = '';
    data.entries.forEach(entry => {
        if (entry.pronunciations && entry.pronunciations.length > 0 && !pronVal) {
            pronVal = entry.pronunciations[0].value;
        }
    });

    if (pronVal) {
        const pronSpan = document.createElement('span');
        pronSpan.className = 'phonetics-label';
        pronSpan.innerText = `/${pronVal}/`;
        mainInfo.appendChild(pronSpan);
    }
    header.appendChild(mainInfo);

    // Controls Row
    const controls = document.createElement('div');
    controls.className = 'controls-row';
    const audioBtn = document.createElement('button');
    audioBtn.className = 'audio-btn';
    audioBtn.innerText = 'Listen';
    audioBtn.onclick = () => speakWord(data.word);
    controls.appendChild(audioBtn);
    header.appendChild(controls);

    resultsArea.appendChild(header);
    updateStarUI();

    // Iterate Lexical entries (POS classifications)
    data.entries.forEach(entry => {
        const entryBlock = document.createElement('article');
        entryBlock.className = 'lexical-entry';

        const posBanner = document.createElement('div');
        posBanner.className = 'pos-banner';
        
        let label = mapPOS(entry.part_of_speech);
        if (entry.forms && entry.forms.length > 0) {
            label += ` — (${entry.forms.join(', ')})`;
        }
        posBanner.innerText = label;
        entryBlock.appendChild(posBanner);

        // Render Senses
        entry.senses.forEach(sense => {
            const senseItem = document.createElement('div');
            senseItem.className = 'sense-item';

            // Definition text
            const defText = document.createElement('div');
            defText.className = 'definition-text';
            defText.innerText = sense.definition.join('; ');
            senseItem.appendChild(defText);

            // Synonym tags list
            if (sense.members && sense.members.length > 1) {
                const synDiv = document.createElement('div');
                synDiv.className = 'definition-synonyms';
                
                sense.members.forEach(member => {
                    if (member.toLowerCase() !== data.word.toLowerCase()) {
                        const tag = document.createElement('span');
                        tag.className = 'synonym-tag';
                        tag.innerText = member.toLowerCase();
                        tag.onclick = () => performSearch(member);
                        synDiv.appendChild(tag);
                    }
                });

                if (synDiv.children.length > 0) {
                    senseItem.appendChild(synDiv);
                }
            }

            // Verb subcategorization frames
            if (sense.sentence_frames && sense.sentence_frames.length > 0) {
                const quotesList = document.createElement('div');
                quotesList.className = 'quotations-list';
                
                sense.sentence_frames.forEach(frame => {
                    const line = document.createElement('p');
                    line.className = 'quote-line';
                    line.innerText = `usage frame: ${frame}`;
                    quotesList.appendChild(line);
                });
                senseItem.appendChild(quotesList);
            }

            // Usage examples
            if (sense.examples && sense.examples.length > 0) {
                const quotesList = document.createElement('div');
                quotesList.className = 'quotations-list';

                sense.examples.forEach(ex => {
                    const line = document.createElement('p');
                    line.className = 'quote-line';
                    line.innerText = `“${ex.text}”`;
                    
                    if (ex.source) {
                        const attrib = document.createElement('span');
                        attrib.className = 'quote-attrib';
                        attrib.innerText = `— ${ex.source}`;
                        line.appendChild(attrib);
                    }
                    quotesList.appendChild(line);
                });
                senseItem.appendChild(quotesList);
            }

            // Sentences from entry level
            if (sense.example_sentences && sense.example_sentences.length > 0) {
                const quotesList = document.createElement('div');
                quotesList.className = 'quotations-list';

                sense.example_sentences.forEach(ex => {
                    const line = document.createElement('p');
                    line.className = 'quote-line';
                    line.innerText = `“${ex}”`;
                    quotesList.appendChild(line);
                });
                senseItem.appendChild(quotesList);
            }

            // Collapsible semantic relations accordion
            if (sense.relations && sense.relations.length > 0) {
                const accordion = document.createElement('div');
                accordion.className = 'relations-accordion';

                sense.relations.forEach((rel, rIdx) => {
                    const block = document.createElement('div');
                    block.className = 'relation-block';
                    
                    const trigger = document.createElement('button');
                    trigger.className = 'relation-trigger';
                    
                    const totalItems = rel.synsets ? rel.synsets.length : rel.senses.length;
                    
                    trigger.innerHTML = `
                        <span>${rel.relation.replace(/_/g, ' ')} (${totalItems})</span>
                        <span class="toggle-icon">[+]</span>`;
                    
                    trigger.onclick = () => toggleRelation(block, trigger);
                    block.appendChild(trigger);

                    const panel = document.createElement('div');
                    panel.className = 'relation-panel';

                    // Parse Synset targets
                    if (rel.synsets) {
                        rel.synsets.forEach(ts => {
                            const card = document.createElement('div');
                            card.className = 'related-card';

                            const wordRow = document.createElement('div');
                            wordRow.className = 'related-words-row';
                            
                            ts.members.forEach((member, mIdx) => {
                                const link = document.createElement('a');
                                link.className = 'related-word-link';
                                link.innerText = member.toLowerCase();
                                link.onclick = () => performSearch(member);
                                wordRow.appendChild(link);

                                if (mIdx < ts.members.length - 1) {
                                    wordRow.appendChild(document.createTextNode(', '));
                                }
                            });
                            card.appendChild(wordRow);

                            const rdef = document.createElement('p');
                            rdef.className = 'related-definition';
                            rdef.innerText = ts.definition.join('; ');
                            card.appendChild(rdef);
                            
                            panel.appendChild(card);
                        });
                    }

                    // Parse Sense targets
                    if (rel.senses) {
                        rel.senses.forEach(ts => {
                            const card = document.createElement('div');
                            card.className = 'related-card';

                            const wordRow = document.createElement('div');
                            wordRow.className = 'related-words-row';
                            const link = document.createElement('a');
                            link.className = 'related-word-link';
                            link.innerText = ts.lemma.toLowerCase();
                            link.onclick = () => performSearch(ts.lemma);
                            wordRow.appendChild(link);
                            card.appendChild(wordRow);

                            if (ts.definition && ts.definition.length > 0) {
                                const rdef = document.createElement('p');
                                rdef.className = 'related-definition';
                                rdef.innerText = ts.definition.join('; ');
                                card.appendChild(rdef);
                            }
                            panel.appendChild(card);
                        });
                    }

                    block.appendChild(panel);
                    accordion.appendChild(block);
                });

                senseItem.appendChild(accordion);
            }

            // Wikidata / ILI metadata rows
            if (sense.ili || (sense.wikidata && sense.wikidata.length > 0)) {
                const metadata = document.createElement('div');
                metadata.className = 'item-metadata';

                if (sense.ili) {
                    metadata.innerHTML += `<span>ili: ${sense.ili}</span>`;
                }

                if (sense.wikidata && sense.wikidata.length > 0) {
                    const links = sense.wikidata.map(w => 
                        `<a href="https://www.wikidata.org/wiki/${w}" target="_blank">${w}</a>`
                    ).join(', ');
                    metadata.innerHTML += `<span>wikidata: ${links}</span>`;
                }

                senseItem.appendChild(metadata);
            }

            entryBlock.appendChild(senseItem);
        });

        resultsArea.appendChild(entryBlock);
    });
}

function toggleRelation(block, trigger) {
    const isExpanded = block.classList.contains('expanded');
    block.classList.toggle('expanded');
    
    const icon = trigger.querySelector('.toggle-icon');
    if (icon) {
        icon.innerText = isExpanded ? '[+]' : '[-]';
    }
}

// Speak pronunciations via Browser SpeechSynthesis
function speakWord(text) {
    if ('speechSynthesis' in window) {
        // Cancel active audio
        window.speechSynthesis.cancel();
        
        const utterance = new SpeechSynthesisUtterance(text);
        utterance.lang = 'en-US';
        window.speechSynthesis.speak(utterance);
    } else {
        alert("Speech synthesis is not supported by your browser.");
    }
}

// Parts of Speech mapper
// POS codes from WordNet format: n (noun), v (verb), a (adj), r (adverb), s (satellite adj)
function mapPOS(pos) {
    switch (pos) {
        case 'n': return 'noun';
        case 'v': return 'verb';
        case 'a': return 'adjective';
        case 'r': return 'adverb';
        case 's': return 'adjective satellite';
        default: return pos;
    }
}

// HTML escape helper
function escapeHTML(str) {
    return str
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

// Swagger Playground initialization
function initSwagger() {
    if (swaggerInitialized) return;
    
    SwaggerUIBundle({
        url: "/api/v1/openapi.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
        plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "BaseLayout"
    });
    
    swaggerInitialized = true;
}
