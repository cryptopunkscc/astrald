// Command ws-client serves a static web page that lets a browser act as a generic
// apphost WebSocket client — the browser equivalent of astral-query.
//
// Usage:
//
//	ws-client -listen 127.0.0.1:8627
//
// Then open http://127.0.0.1:8627 in a browser and point it at a running node's
// /.ws endpoint (default ws://127.0.0.1:8624/.ws).
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8627", "address to bind the web app to")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(indexHTML))
	})

	log.Printf("ws-client serving at http://%s", *listen)
	log.Fatal(http.ListenAndServe(*listen, mux))
}

const indexHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>astral ws-client</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
:root {
  --bg: #0d1117;
  --panel: #161b22;
  --border: #30363d;
  --fg: #c9d1d9;
  --muted: #8b949e;
  --accent: #58a6ff;
  --good: #3fb950;
  --warn: #d29922;
  --bad: #f85149;
  --recv: #79c0ff;
  --send: #d2a8ff;
}
* { box-sizing: border-box; }
html, body { height: 100%; margin: 0; }
body {
  background: var(--bg);
  color: var(--fg);
  font: 13px/1.45 ui-monospace, "JetBrains Mono", Menlo, Consolas, monospace;
  display: grid;
  grid-template-columns: 360px 1fr;
  grid-template-rows: 1fr;
}
.sidebar {
  background: var(--panel);
  border-right: 1px solid var(--border);
  padding: 14px;
  overflow-y: auto;
}
.main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
h1 {
  font-size: 13px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--muted);
  margin: 0 0 12px 0;
  font-weight: 600;
}
h2 {
  font-size: 11px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
  margin: 18px 0 8px 0;
  font-weight: 600;
}
fieldset {
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 10px 12px 12px 12px;
  margin: 0 0 12px 0;
}
fieldset legend {
  padding: 0 6px;
  color: var(--muted);
  font-size: 11px;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}
label {
  display: block;
  margin: 8px 0 3px 0;
  color: var(--muted);
  font-size: 11px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
input[type=text], input[type=number], select, textarea {
  width: 100%;
  background: var(--bg);
  color: var(--fg);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 6px 8px;
  font: inherit;
  outline: none;
}
input:focus, select:focus, textarea:focus {
  border-color: var(--accent);
}
textarea { resize: vertical; min-height: 60px; }
.row { display: flex; gap: 8px; }
.row > * { flex: 1; }
button {
  background: var(--bg);
  color: var(--fg);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 6px 12px;
  font: inherit;
  cursor: pointer;
}
button:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
button:disabled { opacity: 0.4; cursor: not-allowed; }
button.primary {
  background: var(--accent);
  color: #0d1117;
  border-color: var(--accent);
  font-weight: 600;
}
button.primary:hover:not(:disabled) { filter: brightness(1.1); color: #0d1117; }
button.danger { color: var(--bad); border-color: var(--bad); }
button.danger:hover:not(:disabled) { background: var(--bad); color: #0d1117; }
.status {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 11px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.dot { width: 8px; height: 8px; border-radius: 50%; background: var(--muted); }
.status.connecting .dot { background: var(--warn); animation: pulse 1s infinite; }
.status.connected .dot { background: var(--good); }
.status.error .dot { background: var(--bad); }
@keyframes pulse { 50% { opacity: 0.3; } }
.host-info {
  font-size: 11px;
  color: var(--muted);
  margin: 4px 0 0 0;
  word-break: break-all;
}
.host-info code { color: var(--fg); }

.log-toolbar {
  background: var(--panel);
  border-bottom: 1px solid var(--border);
  padding: 8px 14px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 11px;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.log-toolbar .spacer { flex: 1; }
.log-toolbar label { display: inline-flex; align-items: center; gap: 4px; margin: 0; text-transform: none; }
.log {
  flex: 1;
  overflow-y: auto;
  padding: 8px 14px;
  background: var(--bg);
}
.entry {
  display: grid;
  grid-template-columns: 70px 60px 1fr;
  gap: 10px;
  padding: 4px 0;
  border-bottom: 1px solid #1a1f26;
  align-items: start;
}
.entry .time { color: var(--muted); font-size: 11px; padding-top: 2px; }
.entry .dir { font-size: 10px; padding-top: 2px; letter-spacing: 0.06em; }
.entry .dir.recv { color: var(--recv); }
.entry .dir.send { color: var(--send); }
.entry .dir.info { color: var(--muted); }
.entry .dir.err { color: var(--bad); }
.entry .body { word-break: break-all; }
.entry .body pre {
  margin: 0;
  font: inherit;
  white-space: pre-wrap;
  word-break: break-all;
}
.type-tag {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 3px;
  background: #1f2937;
  color: var(--accent);
  font-size: 11px;
  margin-right: 8px;
}

.composer {
  background: var(--panel);
  border-top: 1px solid var(--border);
  padding: 10px 14px;
}
.composer textarea {
  font: inherit;
  min-height: 70px;
}
.composer .row { margin-top: 6px; align-items: center; }
.composer .row .spacer { flex: 1; }
.composer .row button { flex: 0 0 auto; }
.composer .hint { color: var(--muted); font-size: 11px; }

.muted { color: var(--muted); }
.kvp { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; }
</style>
</head>
<body>

<aside class="sidebar">
  <h1>astral ws-client</h1>

  <fieldset>
    <legend>connection</legend>
    <label>websocket url</label>
    <input id="url" type="text" value="ws://127.0.0.1:8624/.ws">
    <label>mode</label>
    <select id="mode">
      <option value="astral.json.v1">json (text frames)</option>
      <option value="astral.binary.v1">binary (raw frames)</option>
    </select>
    <label>access token (optional)</label>
    <input id="token" type="text" placeholder="leave blank for anonymous">
    <div class="row" style="margin-top: 10px;">
      <button id="connectBtn" class="primary">connect</button>
      <button id="disconnectBtn" class="danger" disabled>disconnect</button>
    </div>
    <div id="status" class="status" style="margin-top: 10px;">
      <span class="dot"></span><span id="statusText">disconnected</span>
    </div>
    <div id="hostInfo" class="host-info"></div>
  </fieldset>

  <fieldset>
    <legend>route query</legend>
    <label>target identity (blank = host)</label>
    <input id="target" type="text" placeholder="host">
    <label>caller identity (blank = guest)</label>
    <input id="caller" type="text" placeholder="guest">
    <label>query string</label>
    <input id="query" type="text" placeholder="user.info" value="apphost.list_tokens">
    <label>zone</label>
    <div class="kvp">
      <label style="text-transform: none; margin: 0;"><input type="checkbox" id="zd" checked> device</label>
      <label style="text-transform: none; margin: 0;"><input type="checkbox" id="zv" checked> virtual</label>
      <label style="text-transform: none; margin: 0;"><input type="checkbox" id="zn" checked> network</label>
    </div>
    <label>filters (one per line, optional)</label>
    <textarea id="filters" placeholder="filter-name"></textarea>
    <div class="row" style="margin-top: 10px;">
      <button id="sendQueryBtn" disabled>send route_query</button>
    </div>
  </fieldset>

  <fieldset>
    <legend>raw send</legend>
    <p class="muted" style="margin: 0 0 6px 0; font-size: 11px;">
      send a hand-built astral.Object as JSON envelope, or raw bytes (binary mode, hex).
    </p>
    <button id="sendRawBtn" disabled>send composer</button>
  </fieldset>
</aside>

<section class="main">
  <div class="log-toolbar">
    <span>log</span>
    <label><input type="checkbox" id="autoScroll" checked> autoscroll</label>
    <span class="spacer"></span>
    <button id="clearBtn">clear</button>
  </div>
  <div id="log" class="log"></div>
  <div class="composer">
    <textarea id="composer" placeholder='in json mode: {"Type":"mod.apphost.auth_token_msg","Object":{"Token":"abc"}}
in binary mode: hex bytes like "00 0a 01 02 03"
once a query is accepted, sends are raw frames forwarded to the responder.'></textarea>
    <div class="row">
      <span class="hint" id="composerHint">json envelope</span>
      <span class="spacer"></span>
      <button id="composerSendBtn" disabled>send</button>
    </div>
  </div>
</section>

<script>
'use strict';

const $ = (id) => document.getElementById(id);

let ws = null;
let mode = 'astral.json.v1';
let queryAccepted = false;

const els = {
  url: $('url'),
  mode: $('mode'),
  token: $('token'),
  connectBtn: $('connectBtn'),
  disconnectBtn: $('disconnectBtn'),
  status: $('status'),
  statusText: $('statusText'),
  hostInfo: $('hostInfo'),
  target: $('target'),
  caller: $('caller'),
  query: $('query'),
  zd: $('zd'), zv: $('zv'), zn: $('zn'),
  filters: $('filters'),
  sendQueryBtn: $('sendQueryBtn'),
  sendRawBtn: $('sendRawBtn'),
  log: $('log'),
  autoScroll: $('autoScroll'),
  clearBtn: $('clearBtn'),
  composer: $('composer'),
  composerHint: $('composerHint'),
  composerSendBtn: $('composerSendBtn'),
};

let hostID = null;
let guestID = null;

// ---- logging ----

function log(dir, body, type) {
  const e = document.createElement('div');
  e.className = 'entry';
  const t = new Date();
  const time = t.toTimeString().slice(0,8) + '.' + String(t.getMilliseconds()).padStart(3,'0');
  e.innerHTML = '<div class="time"></div><div class="dir"></div><div class="body"></div>';
  e.querySelector('.time').textContent = time;
  e.querySelector('.dir').textContent = dir.toUpperCase();
  e.querySelector('.dir').classList.add(dir);
  const bodyEl = e.querySelector('.body');
  if (type) {
    const tag = document.createElement('span');
    tag.className = 'type-tag';
    tag.textContent = type;
    bodyEl.appendChild(tag);
  }
  if (typeof body === 'string') {
    const pre = document.createElement('pre');
    pre.textContent = body;
    bodyEl.appendChild(pre);
  } else if (body !== undefined) {
    const pre = document.createElement('pre');
    pre.textContent = JSON.stringify(body, null, 2);
    bodyEl.appendChild(pre);
  }
  els.log.appendChild(e);
  if (els.autoScroll.checked) els.log.scrollTop = els.log.scrollHeight;
}

els.clearBtn.addEventListener('click', () => { els.log.innerHTML = ''; });

// ---- status ----

function setStatus(s, text) {
  els.status.className = 'status ' + s;
  els.statusText.textContent = text;
}

// ---- helpers ----

function newNonce() {
  const b = crypto.getRandomValues(new Uint8Array(8));
  return [...b].map(x => x.toString(16).padStart(2, '0')).join('');
}

// astral.Zone marshals as a string of characters: d=device, v=virtual, n=network.
function zoneString() {
  return (els.zd.checked ? 'd' : '') + (els.zv.checked ? 'v' : '') + (els.zn.checked ? 'n' : '');
}

function filtersList() {
  return els.filters.value.split('\n').map(s => s.trim()).filter(Boolean);
}

function hexToBytes(s) {
  const cleaned = s.replace(/[\s,;:]+/g, '').toLowerCase();
  if (!/^[0-9a-f]*$/.test(cleaned) || cleaned.length % 2) {
    throw new Error('invalid hex');
  }
  const out = new Uint8Array(cleaned.length / 2);
  for (let i = 0; i < out.length; i++) {
    out[i] = parseInt(cleaned.substr(i*2, 2), 16);
  }
  return out;
}

function bytesToHexAscii(bytes) {
  const rows = [];
  for (let i = 0; i < bytes.length; i += 16) {
    const slice = bytes.subarray(i, i + 16);
    const hex = [...slice].map(b => b.toString(16).padStart(2, '0')).join(' ');
    const ascii = [...slice].map(b => (b >= 32 && b < 127) ? String.fromCharCode(b) : '.').join('');
    rows.push(i.toString(16).padStart(4, '0') + '  ' + hex.padEnd(48, ' ') + '  ' + ascii);
  }
  return rows.join('\n');
}

// ---- frame send/receive ----

function sendJSONObject(type, object) {
  const env = { Type: type, Object: object };
  const text = JSON.stringify(env);
  ws.send(text);
  log('send', env, type);
}

function sendBinaryRaw(bytes) {
  ws.send(bytes);
  log('send', bytesToHexAscii(bytes), 'binary ' + bytes.length + 'B');
}

function handleJSONFrame(text) {
  let env;
  try { env = JSON.parse(text); } catch (e) {
    log('err', 'invalid JSON: ' + text);
    return;
  }
  const { Type, Object: obj } = env;
  log('recv', obj, Type);

  if (queryAccepted) return; // post-accept: just log; semantics are app-defined

  switch (Type) {
    case 'mod.apphost.host_info_msg':
      hostID = obj && obj.Identity;
      els.hostInfo.innerHTML = 'host: <code>' + (obj.Alias || '(no alias)') + '</code><br>id: <code>' + (hostID || '?') + '</code>';
      setStatus('connected', 'connected');
      if (els.token.value.trim()) {
        sendJSONObject('mod.apphost.auth_token_msg', { Token: els.token.value.trim() });
      } else {
        // anonymous: enable query immediately
        els.sendQueryBtn.disabled = false;
        els.sendRawBtn.disabled = false;
      }
      break;

    case 'mod.apphost.auth_success_msg':
      guestID = obj && obj.GuestID;
      els.hostInfo.innerHTML += '<br>guest: <code>' + (guestID || '?') + '</code>';
      els.sendQueryBtn.disabled = false;
      els.sendRawBtn.disabled = false;
      break;

    case 'mod.apphost.error_msg':
      log('err', 'host error: ' + (obj && obj.Code));
      break;

    case 'mod.apphost.query_accepted_msg':
      queryAccepted = true;
      log('info', 'query accepted — connection is now a raw stream', 'state');
      els.composerHint.textContent = 'post-accept: payload to responder';
      els.sendQueryBtn.disabled = true;
      break;

    case 'mod.apphost.query_rejected_msg':
      log('err', 'query rejected, code ' + (obj && obj.Code));
      break;
  }
}

function handleBinaryFrame(buf) {
  const bytes = new Uint8Array(buf);
  log('recv', bytesToHexAscii(bytes), 'binary ' + bytes.length + 'B');
}

// ---- connect / disconnect ----

els.connectBtn.addEventListener('click', () => {
  if (ws) return;
  mode = els.mode.value;
  queryAccepted = false;
  hostID = null;
  guestID = null;
  els.hostInfo.textContent = '';
  els.sendQueryBtn.disabled = true;
  els.sendRawBtn.disabled = true;

  let url;
  try { url = els.url.value.trim(); }
  catch (e) { log('err', 'bad url'); return; }
  if (!url) { log('err', 'empty url'); return; }

  setStatus('connecting', 'connecting...');
  log('info', url + ' subprotocol=' + mode, 'connecting');

  try {
    ws = new WebSocket(url, mode);
  } catch (e) {
    log('err', String(e));
    setStatus('error', 'error');
    ws = null;
    return;
  }
  if (mode === 'astral.binary.v1') ws.binaryType = 'arraybuffer';

  els.connectBtn.disabled = true;
  els.disconnectBtn.disabled = false;
  els.composerSendBtn.disabled = false;
  updateComposerHint();

  ws.onopen = () => {
    log('info', 'socket open, awaiting host_info_msg', 'open');
  };
  ws.onmessage = (ev) => {
    if (typeof ev.data === 'string') handleJSONFrame(ev.data);
    else handleBinaryFrame(ev.data);
  };
  ws.onerror = () => {
    log('err', 'socket error');
  };
  ws.onclose = (ev) => {
    log('info', 'closed code=' + ev.code + ' reason=' + (ev.reason || '(none)'), 'close');
    setStatus('', 'disconnected');
    ws = null;
    els.connectBtn.disabled = false;
    els.disconnectBtn.disabled = true;
    els.sendQueryBtn.disabled = true;
    els.sendRawBtn.disabled = true;
    els.composerSendBtn.disabled = true;
    queryAccepted = false;
  };
});

els.disconnectBtn.addEventListener('click', () => {
  if (ws) ws.close();
});

// ---- send route_query ----

els.sendQueryBtn.addEventListener('click', () => {
  if (!ws) return;
  if (mode !== 'astral.json.v1') {
    log('err', 'route_query helper only works in json mode (binary: build the envelope by hand and use composer)');
    return;
  }
  const target = els.target.value.trim() || hostID;
  const caller = els.caller.value.trim() || null;
  const obj = {
    Nonce: newNonce(),
    Caller: caller,
    Target: target,
    Query: els.query.value.trim(),
    Zone: zoneString(),
    Filters: filtersList().length ? filtersList() : null,
  };
  sendJSONObject('mod.apphost.route_query_msg', obj);
});

// ---- composer ----

function updateComposerHint() {
  if (queryAccepted) {
    els.composerHint.textContent = 'post-accept: ' + (mode === 'astral.json.v1' ? 'JSON object to responder' : 'hex bytes to responder');
  } else if (mode === 'astral.json.v1') {
    els.composerHint.textContent = 'json envelope';
  } else {
    els.composerHint.textContent = 'hex bytes';
  }
}
els.mode.addEventListener('change', updateComposerHint);

els.composerSendBtn.addEventListener('click', () => {
  if (!ws) return;
  const text = els.composer.value.trim();
  if (!text) return;
  try {
    if (mode === 'astral.json.v1') {
      // Try to send as one envelope OR as a bare object (post-accept) — accept either.
      const parsed = JSON.parse(text);
      if (parsed && typeof parsed === 'object' && 'Type' in parsed && 'Object' in parsed) {
        ws.send(JSON.stringify(parsed));
        log('send', parsed.Object, parsed.Type);
      } else {
        // bare object: wrap in envelope using Type field if present, else 'unknown'
        ws.send(text);
        log('send', parsed, '(raw)');
      }
    } else {
      const bytes = hexToBytes(text);
      sendBinaryRaw(bytes);
    }
  } catch (e) {
    log('err', 'send failed: ' + e.message);
  }
});

els.sendRawBtn.addEventListener('click', () => {
  els.composer.focus();
});

// Submit composer on Ctrl/Cmd+Enter.
els.composer.addEventListener('keydown', (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    e.preventDefault();
    els.composerSendBtn.click();
  }
});

log('info', 'ready. configure connection, then click connect.', 'boot');
</script>
</body>
</html>
`
