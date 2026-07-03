// Jalankan: go run ./24-websocket
//
//	WebSocket : ws://localhost:8080/ws     (dua arah)
//	SSE       : http://localhost:8080/events (satu arah)
//
// Buka http://localhost:8080 di dua tab browser untuk mencoba chat.
// Verifikasi otomatis: go test ./24-websocket
package main

import (
	"log"
	"net/http"
)

const page = `<!doctype html><html><body>
<h3>Chat WebSocket</h3>
<input id="msg" placeholder="ketik pesan"><button onclick="send()">Kirim</button>
<ul id="log"></ul>
<script>
const ws = new WebSocket("ws://" + location.host + "/ws");
ws.onmessage = e => { const li=document.createElement('li'); li.textContent=e.data; document.getElementById('log').appendChild(li); };
function send(){ const m=document.getElementById('msg'); ws.send(m.value); m.value=''; }
</script></body></html>`

func main() {
	hub := NewHub()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(page))
	})
	mux.HandleFunc("GET /ws", hub.handleWS)
	mux.HandleFunc("GET /events", hub.handleSSE)

	log.Println("buka http://localhost:8080 di dua tab untuk chat realtime")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
