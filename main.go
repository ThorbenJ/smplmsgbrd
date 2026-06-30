package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"text/template"
)

const MAXMSG = 2048

var (
	messages [MAXMSG]string
	// head and tail are absolute, ever-increasing counters.
	// The ring-buffer slot for index i is messages[i % MAXMSG].
	// head == tail means empty; tail - head == MAXMSG means full.
	head int
	tail int
	mu   sync.RWMutex
)

type pageData struct {
	CSS string
	JS  string
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/api/send_msg", sendMsgHandler)
	mux.HandleFunc("/api/peek_msgs", peekMsgsHandler)
	mux.HandleFunc("/api/fetch_msgs", fetchMsgsHandler)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, pageData{CSS: cssStyle, JS: scriptJS}); err != nil {
		log.Printf("template error: %v", err)
	}
}

// sendMsgHandler accepts POST /api/send_msg with JSON body {"message":"text"}.
func sendMsgHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Message == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	messages[tail%MAXMSG] = body.Message
	tail++
	if tail-head > MAXMSG {
		head++
	}
	mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// peekMsgsHandler returns the current absolute head and tail counters.
func peekMsgsHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	h, t := head, tail
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"head": h, "tail": t})
}

// fetchMsgsHandler returns messages in the absolute half-open range [req_head, req_tail).
// req_head and req_tail must fall within the current [head, tail] window.
func fetchMsgsHandler(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	start, err1 := strconv.Atoi(startStr)
	end, err2 := strconv.Atoi(endStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "invalid start/end", http.StatusBadRequest)
		return
	}

	mu.RLock()
	h, t := head, tail
	mu.RUnlock()

	if start < h || start > t || end < h || end > t || start > end {
		http.Error(w, "start/end out of range", http.StatusBadRequest)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	result := []string{}
	for i := start; i < end; i++ {
		result = append(result, messages[i%MAXMSG])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"messages": result})
}

// ─── Templates ───────────────────────────────────────────────────────────────

const cssStyle = `
  :root {
    --dark:    #0C1526;
    --teal:    #00BFB3;
    --teal-hv: #009E97;
    --bg:      #F5F7FA;
    --surface: #FFFFFF;
    --text:    #343741;
    --muted:   #69707D;
    --border:  #D3DAE6;
  }

  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica Neue', Arial, sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
  }

  /* ── Header ── */
  header {
    background: var(--dark);
    padding: 0 28px;
    height: 56px;
    display: flex;
    align-items: center;
    gap: 10px;
    position: sticky;
    top: 0;
    z-index: 10;
    box-shadow: 0 2px 8px rgba(0,0,0,.35);
  }

  .logo {
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -.3px;
    color: #fff;
  }
  .logo span { color: var(--teal); }

  .header-sub {
    font-size: 12px;
    color: #7a8898;
    margin-left: 4px;
    padding-left: 12px;
    border-left: 1px solid #2e3f56;
    line-height: 1;
    padding-top: 2px;
  }

  /* ── Compose bar ── */
  .compose {
    background: var(--surface);
    border-bottom: 1px solid var(--border);
    padding: 18px 24px;
    display: flex;
    justify-content: center;
    box-shadow: 0 1px 4px rgba(0,0,0,.06);
  }

  .compose-inner {
    display: flex;
    gap: 10px;
    width: 100%;
    max-width: 760px;
  }

  .compose-inner input {
    flex: 1;
    height: 40px;
    padding: 0 14px;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 14px;
    color: var(--text);
    outline: none;
    transition: border-color .15s, box-shadow .15s;
    background: var(--surface);
  }
  .compose-inner input:focus {
    border-color: var(--teal);
    box-shadow: 0 0 0 3px rgba(0,191,179,.18);
  }
  .compose-inner input::placeholder { color: #a8b0bc; }

  .compose-inner button {
    height: 40px;
    padding: 0 24px;
    background: var(--teal);
    color: #fff;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: background .15s;
    white-space: nowrap;
  }
  .compose-inner button:hover  { background: var(--teal-hv); }
  .compose-inner button:active { background: #008880; }

  /* ── Feed ── */
  .feed-wrap {
    max-width: 760px;
    margin: 24px auto;
    padding: 0 24px 48px;
  }

  .feed-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }

  .feed-label {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: .08em;
    color: var(--muted);
  }

  .pulse {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--teal);
    margin-right: 7px;
    vertical-align: middle;
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50%       { opacity: .35; }
  }

  .msg-count {
    font-size: 12px;
    color: var(--muted);
  }

  #board {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .msg-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px 16px;
    font-size: 14px;
    line-height: 1.55;
    word-break: break-word;
    box-shadow: 0 1px 3px rgba(0,0,0,.04);
    animation: slide-in .18s ease;
  }

  @keyframes slide-in {
    from { opacity: 0; transform: translateY(-6px); }
    to   { opacity: 1; transform: translateY(0); }
  }

  .empty-state {
    text-align: center;
    padding: 56px 0;
    color: var(--muted);
    font-size: 14px;
    border: 2px dashed var(--border);
    border-radius: 8px;
  }
  .empty-state p { margin-top: 6px; font-size: 12px; color: #a8b0bc; }

  .eviction-notice {
    text-align: center;
    font-size: 12px;
    color: var(--muted);
    padding: 8px;
    margin-bottom: 4px;
    border-radius: 4px;
    background: #fffae6;
    border: 1px solid #f5d76e;
  }
`

const scriptJS = `
  const POLL_MS = 3000;

  let knownHead = 0;
  let knownTail = 0;
  let msgCount  = 0;
  let ready     = false;

  async function apiPeek() {
    const r = await fetch('/api/peek_msgs');
    if (!r.ok) throw new Error('peek ' + r.status);
    return r.json();
  }

  async function apiFetch(start, end) {
    if (start === end) return [];
    const r = await fetch('/api/fetch_msgs?start=' + start + '&end=' + end);
    if (!r.ok) throw new Error('fetch ' + r.status);
    const d = await r.json();
    return d.messages || [];
  }

  function setCount(n) {
    msgCount = n;
    document.getElementById('msg-count').textContent =
      n === 0 ? '' : n + ' message' + (n === 1 ? '' : 's');
  }

  function appendCards(msgs, scrollToNew) {
    const board = document.getElementById('board');
    const empty = document.getElementById('empty-state');
    if (empty && msgs.length > 0) empty.remove();

    for (const text of msgs) {
      const div = document.createElement('div');
      div.className = 'msg-card';
      div.textContent = text;
      board.appendChild(div);
    }
    setCount(msgCount + msgs.length);

    if (scrollToNew && msgs.length > 0) {
      window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' });
    }
  }

  function insertEvictionNotice() {
    const board = document.getElementById('board');
    const notice = document.createElement('div');
    notice.className = 'eviction-notice';
    notice.textContent = '— older messages were removed from the buffer —';
    board.prepend(notice);
  }

  async function init() {
    try {
      const { head, tail } = await apiPeek();
      knownHead = head;
      knownTail = tail;
      const msgs = await apiFetch(head, tail);
      if (msgs.length > 0) appendCards(msgs, false);
      ready = true;
    } catch (err) {
      console.error('init:', err);
    }
  }

  async function poll() {
    if (!ready) return;
    try {
      const { head, tail } = await apiPeek();

      if (tail === knownTail && head === knownHead) return;

      if (head !== knownHead) {
        // The oldest messages were evicted — reload the full current board.
        const board = document.getElementById('board');
        board.innerHTML = '';
        setCount(0);
        insertEvictionNotice();
        knownTail = head; // fetch from new head forward
      }

      if (tail !== knownTail) {
        const msgs = await apiFetch(knownTail, tail);
        appendCards(msgs, true);
      }

      knownHead = head;
      knownTail = tail;
    } catch (err) {
      console.error('poll:', err);
    }
  }

  async function sendMessage() {
    const input = document.getElementById('msg-input');
    const text  = input.value.trim();
    if (!text) return;

    input.disabled = true;
    try {
      const r = await fetch('/api/send_msg', {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({ message: text }),
      });
      if (r.ok) {
        input.value = '';
        await poll();
      }
    } catch (err) {
      console.error('send:', err);
    } finally {
      input.disabled = false;
      input.focus();
    }
  }

  document.addEventListener('DOMContentLoaded', () => {
    init();
    setInterval(poll, POLL_MS);

    document.getElementById('msg-input').addEventListener('keydown', e => {
      if (e.key === 'Enter') sendMessage();
    });
  });
`

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Simple Message Board</title>
  <style>{{.CSS}}</style>
</head>
<body>

  <header>
    <div class="logo">msg<span>board</span></div>
    <span class="header-sub">shared live feed</span>
  </header>

  <div class="compose">
    <div class="compose-inner">
      <input id="msg-input" type="text" placeholder="Type a message and press Enter or Send…" maxlength="500" autofocus />
      <button onclick="sendMessage()">Send</button>
    </div>
  </div>

  <div class="feed-wrap">
    <div class="feed-header">
      <span class="feed-label"><span class="pulse"></span>Live feed</span>
      <span id="msg-count" class="msg-count"></span>
    </div>
    <div id="board">
      <div id="empty-state" class="empty-state">
        No messages yet.<p>Be the first to say something.</p>
      </div>
    </div>
  </div>

  <script>{{.JS}}</script>
</body>
</html>`
