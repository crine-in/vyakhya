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
	"net/http"
)

// handleLandingPage serves the premium search interface and docs.
func (s *Server) handleLandingPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"page not found"}`))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(landingPageHTML))
}

const landingPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Vyakhya — High-Performance WordNet Dictionary API</title>
    <meta name="description" content="A ultra-fast, memory-efficient REST API and interactive interface for English WordNet.">
    <!-- Google Fonts -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=Outfit:wght@400;500;600;700;800&family=Fira+Code:wght@400;500&display=swap" rel="stylesheet">
    <!-- Swagger UI CSS -->
    <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.9.0/swagger-ui.css" />
    <style>
        :root {
            --bg-color: #060b13;
            --card-bg: rgba(13, 20, 35, 0.6);
            --card-border: rgba(255, 255, 255, 0.08);
            --accent-cyan: #06b6d4;
            --accent-purple: #8b5cf6;
            --accent-green: #10b981;
            --text-primary: #f8fafc;
            --text-secondary: #94a3b8;
            --font-outfit: 'Outfit', sans-serif;
            --font-inter: 'Inter', sans-serif;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            background-color: var(--bg-color);
            background-image: 
                radial-gradient(at 0% 0%, rgba(139, 92, 246, 0.15) 0px, transparent 50%),
                radial-gradient(at 100% 100%, rgba(6, 182, 212, 0.15) 0px, transparent 50%);
            background-attachment: fixed;
            color: var(--text-primary);
            font-family: var(--font-inter);
            min-height: 100vh;
            overflow-x: hidden;
            line-height: 1.6;
        }

        header {
            max-width: 1400px;
            margin: 0 auto;
            padding: 2rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .logo-area {
            display: flex;
            align-items: center;
            gap: 0.75rem;
        }

        .logo-icon {
            width: 40px;
            height: 40px;
            background: linear-gradient(135deg, var(--accent-purple), var(--accent-cyan));
            border-radius: 12px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-family: var(--font-outfit);
            font-weight: 800;
            font-size: 1.5rem;
            color: #fff;
            box-shadow: 0 0 20px rgba(139, 92, 246, 0.4);
        }

        .logo-title {
            font-family: var(--font-outfit);
            font-weight: 700;
            font-size: 1.8rem;
            background: linear-gradient(135deg, #fff, #94a3b8);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            letter-spacing: -0.5px;
        }

        nav {
            display: flex;
            gap: 1.5rem;
        }

        .nav-btn {
            background: none;
            border: none;
            color: var(--text-secondary);
            font-family: var(--font-outfit);
            font-weight: 600;
            font-size: 1rem;
            cursor: pointer;
            padding: 0.5rem 1rem;
            border-radius: 8px;
            transition: all 0.3s ease;
        }

        .nav-btn.active, .nav-btn:hover {
            color: var(--text-primary);
            background: rgba(255, 255, 255, 0.05);
        }

        main {
            max-width: 1400px;
            margin: 0 auto;
            padding: 0 2rem 4rem 2rem;
        }

        .tab-content {
            display: none;
            animation: fadeIn 0.4s ease-out forwards;
        }

        .tab-content.active {
            display: block;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        /* Dashboard Grid Layout */
        .dashboard-grid {
            display: grid;
            grid-template-columns: 2fr 1fr;
            gap: 2rem;
        }

        @media (max-width: 1024px) {
            .dashboard-grid {
                grid-template-columns: 1fr;
            }
        }

        /* Glassmorphic Panel */
        .glass-panel {
            background: var(--card-bg);
            border: 1px solid var(--card-border);
            border-radius: 20px;
            padding: 2rem;
            backdrop-filter: blur(16px) saturate(180%);
            -webkit-backdrop-filter: blur(16px) saturate(180%);
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.37);
        }

        /* Search Section */
        .search-container {
            position: relative;
            margin-bottom: 2rem;
        }

        .search-input {
            width: 100%;
            padding: 1.25rem 1.5rem;
            padding-right: 4rem;
            background: rgba(10, 15, 30, 0.8);
            border: 2px solid var(--card-border);
            border-radius: 16px;
            color: var(--text-primary);
            font-family: var(--font-inter);
            font-size: 1.2rem;
            outline: none;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            box-shadow: inset 0 2px 4px rgba(0,0,0,0.5);
        }

        .search-input:focus {
            border-color: var(--accent-purple);
            box-shadow: 0 0 20px rgba(139, 92, 246, 0.2), inset 0 2px 4px rgba(0,0,0,0.5);
        }

        .search-icon-btn {
            position: absolute;
            right: 1.25rem;
            top: 50%;
            transform: translateY(-50%);
            background: none;
            border: none;
            color: var(--text-secondary);
            cursor: pointer;
            transition: color 0.3s ease;
        }

        .search-icon-btn:hover {
            color: var(--accent-cyan);
        }

        /* Stats Cards */
        .stats-panel {
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
        }

        .stats-header {
            font-family: var(--font-outfit);
            font-size: 1.4rem;
            font-weight: 700;
            margin-bottom: 0.5rem;
            background: linear-gradient(135deg, var(--accent-cyan), var(--accent-purple));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        .stat-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 1rem 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        }

        .stat-item:last-child {
            border: none;
        }

        .stat-label {
            color: var(--text-secondary);
            font-size: 0.95rem;
        }

        .stat-val {
            font-family: 'Fira Code', monospace;
            font-weight: 600;
            color: var(--text-primary);
        }

        /* Results Display */
        .results-box {
            min-height: 300px;
            display: flex;
            flex-direction: column;
            gap: 2rem;
        }

        .placeholder-text {
            color: var(--text-secondary);
            text-align: center;
            margin-top: 5rem;
            font-size: 1.1rem;
        }

        .word-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            border-bottom: 2px solid rgba(255,255,255,0.05);
            padding-bottom: 1.5rem;
        }

        .word-title-wrap h1 {
            font-family: var(--font-outfit);
            font-size: 3rem;
            font-weight: 800;
            letter-spacing: -1px;
            line-height: 1.1;
        }

        .speak-btn {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid var(--card-border);
            border-radius: 50%;
            width: 44px;
            height: 44px;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            color: var(--text-primary);
            transition: all 0.3s ease;
        }

        .speak-btn:hover {
            background: var(--accent-cyan);
            border-color: var(--accent-cyan);
            box-shadow: 0 0 15px rgba(6, 182, 212, 0.4);
            transform: scale(1.05);
        }

        .pos-badge {
            background: rgba(139, 92, 246, 0.15);
            border: 1px solid rgba(139, 92, 246, 0.3);
            color: #a78bfa;
            font-family: var(--font-outfit);
            font-weight: 700;
            padding: 0.25rem 0.75rem;
            border-radius: 20px;
            font-size: 0.85rem;
            text-transform: uppercase;
        }

        .pronunciation {
            color: var(--accent-cyan);
            font-family: 'Fira Code', monospace;
            font-size: 1.05rem;
            margin-top: 0.5rem;
        }

        .entry-block {
            margin-bottom: 2.5rem;
            background: rgba(255, 255, 255, 0.015);
            border-radius: 16px;
            padding: 1.5rem;
            border: 1px solid rgba(255, 255, 255, 0.03);
        }

        .entry-block-title {
            display: flex;
            align-items: center;
            gap: 1rem;
            margin-bottom: 1.5rem;
        }

        .sense-card {
            background: rgba(0, 0, 0, 0.2);
            border: 1px solid rgba(255,255,255,0.03);
            border-radius: 12px;
            padding: 1.25rem;
            margin-bottom: 1.25rem;
            transition: border-color 0.3s ease;
        }

        .sense-card:hover {
            border-color: rgba(255,255,255,0.08);
        }

        .sense-def {
            font-size: 1.1rem;
            font-weight: 500;
            color: var(--text-primary);
        }

        .sense-synonyms {
            display: flex;
            flex-wrap: wrap;
            gap: 0.5rem;
            margin-top: 0.75rem;
        }

        .synonym-pill {
            background: rgba(255, 255, 255, 0.04);
            border: 1px solid var(--card-border);
            color: var(--text-secondary);
            font-size: 0.85rem;
            padding: 0.2rem 0.6rem;
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .synonym-pill:hover {
            background: rgba(6, 182, 212, 0.1);
            border-color: var(--accent-cyan);
            color: var(--text-primary);
        }

        .example-list {
            margin-top: 1rem;
            padding-left: 1.25rem;
            border-left: 2px solid rgba(255, 255, 255, 0.1);
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        .example-item {
            font-style: italic;
            color: var(--text-secondary);
            font-size: 0.95rem;
        }

        .example-source {
            font-family: var(--font-outfit);
            font-size: 0.8rem;
            color: rgba(255, 255, 255, 0.4);
            font-style: normal;
            margin-left: 0.5rem;
        }

        .relations-accordion {
            margin-top: 1.25rem;
        }

        .relation-section {
            border-top: 1px solid rgba(255, 255, 255, 0.05);
            padding: 0.75rem 0;
        }

        .relation-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
            font-weight: 600;
            font-size: 0.95rem;
            color: var(--text-secondary);
            text-transform: capitalize;
            user-select: none;
            padding: 0.25rem 0;
        }

        .relation-header:hover {
            color: var(--text-primary);
        }

        .relation-body {
            display: none;
            padding-top: 0.75rem;
            flex-direction: column;
            gap: 0.75rem;
        }

        .relation-section.active .relation-body {
            display: flex;
        }

        .relation-item {
            background: rgba(0, 0, 0, 0.3);
            border-radius: 8px;
            padding: 0.75rem 1rem;
            border: 1px solid rgba(255, 255, 255, 0.02);
        }

        .relation-item-header {
            display: flex;
            gap: 0.5rem;
            flex-wrap: wrap;
            margin-bottom: 0.25rem;
        }

        .relation-link {
            font-weight: 600;
            color: var(--accent-cyan);
            cursor: pointer;
            text-decoration: none;
        }

        .relation-link:hover {
            text-decoration: underline;
        }

        .relation-def {
            font-size: 0.88rem;
            color: var(--text-secondary);
        }

        /* Swagger UI Theme custom overwrites to match dark design */
        #swagger-ui {
            background: var(--card-bg);
            border: 1px solid var(--card-border);
            border-radius: 20px;
            padding: 1.5rem;
            backdrop-filter: blur(16px) saturate(180%);
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.37);
        }

        .swagger-ui .info .title {
            color: var(--text-primary) !important;
            font-family: var(--font-outfit) !important;
        }

        .swagger-ui .info p, .swagger-ui .info li, .swagger-ui .info td, .swagger-ui .info a {
            color: var(--text-secondary) !important;
        }

        .swagger-ui .scheme-container {
            background: rgba(0,0,0,0.3) !important;
            border-radius: 12px;
            border: 1px solid var(--card-border) !important;
        }

        .swagger-ui .opblock {
            background: rgba(0,0,0,0.2) !important;
            border-color: rgba(255, 255, 255, 0.08) !important;
        }

        .swagger-ui .opblock .opblock-summary-method {
            border-radius: 6px !important;
        }

        .swagger-ui .opblock .opblock-summary-path {
            color: var(--text-primary) !important;
        }

        .swagger-ui .opblock .opblock-summary-description {
            color: var(--text-secondary) !important;
        }

        .swagger-ui input[type=text], .swagger-ui select {
            background: rgba(10, 15, 30, 0.8) !important;
            color: var(--text-primary) !important;
            border-color: var(--card-border) !important;
        }

        .swagger-ui .opblock-body pre.microlight {
            background: #0f172a !important;
            border-color: var(--card-border) !important;
            border-radius: 10px;
        }

        .swagger-ui .response-col_status {
            color: var(--text-primary) !important;
        }

        .swagger-ui table thead tr td, .swagger-ui table thead tr th {
            color: var(--text-primary) !important;
            border-bottom: 2px solid rgba(255,255,255,0.08) !important;
        }

        .swagger-ui .tabli button {
            color: var(--text-secondary) !important;
        }
    </style>
</head>
<body>

    <header>
        <div class="logo-area">
            <div class="logo-icon">V</div>
            <div class="logo-title">Vyakhya</div>
        </div>
        <nav>
            <button class="nav-btn active" onclick="switchTab('explorer')">Explorer</button>
            <button class="nav-btn" onclick="switchTab('docs')">API & Playground</button>
        </nav>
    </header>

    <main>
        <!-- Tab 1: Explorer -->
        <div id="tab-explorer" class="tab-content active">
            <div class="dashboard-grid">
                <!-- Search & Results -->
                <div class="glass-panel">
                    <div class="search-container">
                        <input type="text" id="query-input" class="search-input" placeholder="Search for any word (e.g. happy, abash, a-couple-of)..." autocomplete="off" onkeydown="if(event.key==='Enter') performSearch()">
                        <button class="search-icon-btn" onclick="performSearch()">
                            <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                                <circle cx="11" cy="11" r="8"></circle>
                                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                            </svg>
                        </button>
                    </div>

                    <div id="results-container" class="results-box">
                        <div class="placeholder-text">
                            Type a word and hit Enter to explore definitions and relationships
                        </div>
                    </div>
                </div>

                <!-- Stats & Info -->
                <div class="glass-panel stats-panel" style="height: fit-content;">
                    <div class="stats-header">System Metrics</div>
                    <div class="stat-item">
                        <span class="stat-label">Total Words</span>
                        <span class="stat-val" id="stat-words">-</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Total Synsets</span>
                        <span class="stat-val" id="stat-synsets">-</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Total Senses</span>
                        <span class="stat-val" id="stat-senses">-</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Dataset Load Time</span>
                        <span class="stat-val" id="stat-load">-</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Allocated RAM</span>
                        <span class="stat-val" id="stat-alloc">-</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">System RAM</span>
                        <span class="stat-val" id="stat-sys">-</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Server Uptime</span>
                        <span class="stat-val" id="stat-uptime">-</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Tab 2: Swagger Docs -->
        <div id="tab-docs" class="tab-content">
            <div id="swagger-ui"></div>
        </div>
    </main>

    <!-- Swagger UI Scripts -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.9.0/swagger-ui-standalone-preset.js"></script>

    <script>
        // Tab switching logic
        function switchTab(tabId) {
            document.querySelectorAll('.nav-btn').forEach(btn => btn.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));

            if (tabId === 'explorer') {
                document.querySelector('button[onclick="switchTab(\'explorer\')"]').classList.add('active');
                document.getElementById('tab-explorer').classList.add('active');
            } else {
                document.querySelector('button[onclick="switchTab(\'docs\')"]').classList.add('active');
                document.getElementById('tab-docs').classList.add('active');
                initSwagger();
            }
        }

        // Live stats fetcher
        async function fetchStats() {
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
                console.error("Failed to fetch server statistics:", err);
            }
        }

        // Initial stats fetch and cron
        fetchStats();
        setInterval(fetchStats, 5000);

        // Search lookup logic
        async function performSearch(word) {
            const query = word || document.getElementById('query-input').value.trim();
            if (!query) return;
            
            if (word) {
                document.getElementById('query-input').value = word;
            }

            const resultsContainer = document.getElementById('results-container');
            resultsContainer.innerHTML = '<div class="placeholder-text">Searching dictionary index...</div>';

            try {
                const response = await fetch('/api/v1/word/' + encodeURIComponent(query));
                if (!response.ok) {
                    if (response.status === 404) {
                        resultsContainer.innerHTML = '<div class="placeholder-text" style="color: var(--accent-purple);">Word not found in dictionary index.</div>';
                    } else {
                        resultsContainer.innerHTML = '<div class="placeholder-text" style="color: #ef4444;">An error occurred during lookup.</div>';
                    }
                    return;
                }

                const data = await response.json();
                renderResult(data);
            } catch (err) {
                console.error(err);
                resultsContainer.innerHTML = '<div class="placeholder-text" style="color: #ef4444;">Connection failed. Check server status.</div>';
            }
        }

        // Text-to-speech speaker
        function speakWord(text) {
            if ('speechSynthesis' in window) {
                const utterance = new SpeechSynthesisUtterance(text);
                utterance.lang = 'en-US';
                window.speechSynthesis.speak(utterance);
            } else {
                alert("Speech synthesis not supported in this browser.");
            }
        }

        // Accordion toggle helper
        function toggleAccordion(element) {
            const section = element.parentElement;
            section.classList.toggle('active');
            
            const arrow = element.querySelector('.arrow');
            if (arrow) {
                arrow.style.transform = section.classList.contains('active') ? 'rotate(90deg)' : 'rotate(0deg)';
            }
        }

        // Render result helper
        function renderResult(data) {
            const container = document.getElementById('results-container');
            container.innerHTML = '';

            // Word Header
            const header = document.createElement('div');
            header.className = 'word-header';
            
            const titleWrap = document.createElement('div');
            titleWrap.className = 'word-title-wrap';
            
            const h1 = document.createElement('h1');
            h1.innerText = data.word;
            titleWrap.appendChild(h1);

            // Fetch pronunciations if any
            let pronVal = '';
            data.entries.forEach(entry => {
                if (entry.pronunciations && entry.pronunciations.length > 0 && !pronVal) {
                    pronVal = entry.pronunciations[0].value;
                }
            });

            if (pronVal) {
                const pronDiv = document.createElement('div');
                pronDiv.className = 'pronunciation';
                pronDiv.innerText = '/' + pronVal + '/';
                titleWrap.appendChild(pronDiv);
            }
            
            header.appendChild(titleWrap);

            const speakBtn = document.createElement('button');
            speakBtn.className = 'speak-btn';
            speakBtn.innerHTML = '<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">' +
                '<polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"></polygon>' +
                '<path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"></path>' +
                '</svg>';
            speakBtn.onclick = () => speakWord(data.word);
            header.appendChild(speakBtn);
            container.appendChild(header);

            // Entries blocks
            data.entries.forEach(entry => {
                const block = document.createElement('div');
                block.className = 'entry-block';

                const title = document.createElement('div');
                title.className = 'entry-block-title';
                
                const badge = document.createElement('span');
                badge.className = 'pos-badge';
                badge.innerText = mapPOS(entry.part_of_speech);
                title.appendChild(badge);

                if (entry.forms && entry.forms.length > 0) {
                    const formsSpan = document.createElement('span');
                    formsSpan.style.color = 'var(--text-secondary)';
                    formsSpan.style.fontSize = '0.9rem';
                    formsSpan.innerText = 'Other forms: ' + entry.forms.join(', ');
                    title.appendChild(formsSpan);
                }
                block.appendChild(title);

                // Senses
                entry.senses.forEach(sense => {
                    const card = document.createElement('div');
                    card.className = 'sense-card';

                    const def = document.createElement('div');
                    def.className = 'sense-def';
                    def.innerText = sense.definition.join('; ');
                    card.appendChild(def);

                    // Synonyms/Members
                    if (sense.members && sense.members.length > 1) {
                        const synsDiv = document.createElement('div');
                        synsDiv.className = 'sense-synonyms';
                        
                        sense.members.forEach(member => {
                            if (member.toLowerCase() !== data.word.toLowerCase()) {
                                const pill = document.createElement('span');
                                pill.className = 'synonym-pill';
                                pill.innerText = member;
                                pill.onclick = () => performSearch(member);
                                synsDiv.appendChild(pill);
                            }
                        });

                        if (synsDiv.children.length > 0) {
                            card.appendChild(synsDiv);
                        }
                    }

                    // Verb sentence frames
                    if (sense.sentence_frames && sense.sentence_frames.length > 0) {
                        const framesList = document.createElement('div');
                        framesList.className = 'example-list';
                        framesList.style.borderColor = 'var(--accent-purple)';
                        
                        sense.sentence_frames.forEach(frame => {
                            const item = document.createElement('div');
                            item.className = 'example-item';
                            item.style.color = 'var(--text-primary)';
                            item.innerText = '💬 Frame: ' + frame;
                            framesList.appendChild(item);
                        });
                        card.appendChild(framesList);
                    }

                    // Example sentences/usage
                    if (sense.examples && sense.examples.length > 0) {
                        const exList = document.createElement('div');
                        exList.className = 'example-list';
                        
                        sense.examples.forEach(ex => {
                            const item = document.createElement('div');
                            item.className = 'example-item';
                            item.innerText = '“' + ex.text + '”';

                            if (ex.source) {
                                const src = document.createElement('span');
                                src.className = 'example-source';
                                src.innerText = '— ' + ex.source;
                                item.appendChild(src);
                            }

                            exList.appendChild(item);
                        });
                        card.appendChild(exList);
                    }

                    // Example sentences from entries (sent)
                    if (sense.example_sentences && sense.example_sentences.length > 0) {
                        const exList = document.createElement('div');
                        exList.className = 'example-list';
                        
                        sense.example_sentences.forEach(sent => {
                            const item = document.createElement('div');
                            item.className = 'example-item';
                            item.innerText = '“' + sent + '”';
                            exList.appendChild(item);
                        });
                        card.appendChild(exList);
                    }

                    // Relations Accordion
                    if (sense.relations && sense.relations.length > 0) {
                        const accordion = document.createElement('div');
                        accordion.className = 'relations-accordion';

                        // Group relations
                        sense.relations.forEach(rel => {
                            const section = document.createElement('div');
                            section.className = 'relation-section';

                            const header = document.createElement('div');
                            header.className = 'relation-header';
                            header.onclick = () => toggleAccordion(header);
                            const len = rel.synsets ? rel.synsets.length : rel.senses.length;
                            header.innerHTML = '<span>' + rel.relation.replace('_', ' ') + ' (' + len + ')</span>' +
                                '<span class="arrow" style="transition: transform 0.3s ease; display: inline-block;">▶</span>';
                            section.appendChild(header);

                            const body = document.createElement('div');
                            body.className = 'relation-body';

                            // Render synsets
                            if (rel.synsets) {
                                rel.synsets.forEach(ts => {
                                    const item = document.createElement('div');
                                    item.className = 'relation-item';

                                    const itemH = document.createElement('div');
                                    itemH.className = 'relation-item-header';
                                    
                                    ts.members.forEach((m, idx) => {
                                        const link = document.createElement('a');
                                        link.className = 'relation-link';
                                        link.innerText = m;
                                        link.onclick = () => performSearch(m);
                                        itemH.appendChild(link);

                                        if (idx < ts.members.length - 1) {
                                            const separator = document.createTextNode(', ');
                                            itemH.appendChild(separator);
                                        }
                                    });
                                    item.appendChild(itemH);

                                    const def = document.createElement('div');
                                    def.className = 'relation-def';
                                    def.innerText = ts.definition.join('; ');
                                    item.appendChild(def);

                                    body.appendChild(item);
                                });
                            }

                            // Render senses
                            if (rel.senses) {
                                rel.senses.forEach(ts => {
                                    const item = document.createElement('div');
                                    item.className = 'relation-item';

                                    const itemH = document.createElement('div');
                                    itemH.className = 'relation-item-header';
                                    
                                    const link = document.createElement('a');
                                    link.className = 'relation-link';
                                    link.innerText = ts.lemma;
                                    link.onclick = () => performSearch(ts.lemma);
                                    itemH.appendChild(link);
                                    item.appendChild(itemH);

                                    if (ts.definition && ts.definition.length > 0) {
                                        const def = document.createElement('div');
                                        def.className = 'relation-def';
                                        def.innerText = ts.definition.join('; ');
                                        item.appendChild(def);
                                    }

                                    body.appendChild(item);
                                });
                            }

                            section.appendChild(body);
                            accordion.appendChild(section);
                        });

                        card.appendChild(accordion);
                    }

                    card.appendChild(document.createElement('br'));
                    // Add Wikidata and ILI indicators if they exist
                    if (sense.ili || (sense.wikidata && sense.wikidata.length > 0)) {
                        const metaDiv = document.createElement('div');
                        metaDiv.style.display = 'flex';
                        metaDiv.style.gap = '1rem';
                        metaDiv.style.fontSize = '0.75rem';
                        metaDiv.style.color = 'rgba(255,255,255,0.3)';

                        if (sense.ili) {
                            metaDiv.innerHTML += '<span>ILI: ' + sense.ili + '</span>';
                        }
                        if (sense.wikidata && sense.wikidata.length > 0) {
                            const links = sense.wikidata.map(w => 
                                '<a href="https://www.wikidata.org/wiki/' + w + '" target="_blank" style="color: var(--accent-cyan); text-decoration: none;">' + w + '</a>'
                            ).join(', ');
                            metaDiv.innerHTML += '<span>Wikidata: ' + links + '</span>';
                        }
                        card.appendChild(metaDiv);
                    }

                    block.appendChild(card);
                });

                container.appendChild(block);
            });
        }

        function mapPOS(pos) {
            switch(pos) {
                case 'n': return 'Noun';
                case 'v': return 'Verb';
                case 'a': return 'Adjective';
                case 'r': return 'Adverb';
                case 's': return 'Adjective Satellite';
                default: return pos;
            }
        }

        // Initialize Swagger UI
        let swaggerInitialized = false;
        function initSwagger() {
            if (swaggerInitialized) return;
            const ui = SwaggerUIBundle({
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
    </script>
</body>
</html>
`
