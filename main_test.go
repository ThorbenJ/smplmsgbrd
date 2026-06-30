package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// resetBuffer clears the ring buffer between tests.
func resetBuffer() {
	mu.Lock()
	defer mu.Unlock()
	head = 0
	tail = 0
	messages = [MAXMSG]string{}
}

func post(t *testing.T, msg string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"message": msg})
	r := httptest.NewRequest(http.MethodPost, "/api/send_msg", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	sendMsgHandler(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("post %q: got %d", msg, w.Code)
	}
}

func peekMsgs(t *testing.T) (h, tl int) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/api/peek_msgs", nil)
	w := httptest.NewRecorder()
	peekMsgsHandler(w, r)
	var out map[string]int
	json.NewDecoder(w.Body).Decode(&out)
	return out["head"], out["tail"]
}

func fetchMsgs(t *testing.T, start, end int) []string {
	t.Helper()
	url := "/api/fetch_msgs?start=" + itoa(start) + "&end=" + itoa(end)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	fetchMsgsHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("fetch_msgs(%d,%d): status %d", start, end, w.Code)
	}
	var out map[string][]string
	json.NewDecoder(w.Body).Decode(&out)
	return out["messages"]
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func TestEmptyBuffer(t *testing.T) {
	resetBuffer()
	h, tl := peekMsgs(t)
	if h != 0 || tl != 0 {
		t.Fatalf("want head=0 tail=0, got head=%d tail=%d", h, tl)
	}
	msgs := fetchMsgs(t, 0, 0)
	if len(msgs) != 0 {
		t.Fatalf("want 0 messages, got %d", len(msgs))
	}
}

func TestBasicSendAndFetch(t *testing.T) {
	resetBuffer()
	post(t, "alpha")
	post(t, "beta")
	post(t, "gamma")

	h, tl := peekMsgs(t)
	if h != 0 || tl != 3 {
		t.Fatalf("want head=0 tail=3, got head=%d tail=%d", h, tl)
	}

	msgs := fetchMsgs(t, h, tl)
	want := []string{"alpha", "beta", "gamma"}
	for i, m := range want {
		if msgs[i] != m {
			t.Fatalf("msg[%d]: want %q got %q", i, m, msgs[i])
		}
	}
}

func TestEviction(t *testing.T) {
	resetBuffer()

	// Fill MAXMSG-1 slots (one short of full).
	for i := 0; i < MAXMSG-1; i++ {
		post(t, "x")
	}
	h, tl := peekMsgs(t)
	if h != 0 || tl != MAXMSG-1 {
		t.Fatalf("want head=0 tail=%d, got head=%d tail=%d", MAXMSG-1, h, tl)
	}

	// One more message fills the buffer — no eviction yet.
	post(t, "full")
	h, tl = peekMsgs(t)
	if h != 0 || tl != MAXMSG {
		t.Fatalf("want head=0 tail=%d after fill, got head=%d tail=%d", MAXMSG, h, tl)
	}

	// One more forces eviction: head advances.
	post(t, "overflow")
	h, tl = peekMsgs(t)
	if h != 1 || tl != MAXMSG+1 {
		t.Fatalf("want head=1 tail=%d after eviction, got head=%d tail=%d", MAXMSG+1, h, tl)
	}

	// The last inserted message is at absolute index MAXMSG (slot MAXMSG%MAXMSG == 0).
	msgs := fetchMsgs(t, MAXMSG, MAXMSG+1)
	if len(msgs) != 1 || msgs[0] != "overflow" {
		t.Fatalf("want [overflow], got %v", msgs)
	}
}

func TestWrappedFetch(t *testing.T) {
	resetBuffer()

	// Seed the absolute counters so tail wraps past MAXMSG during the test.
	mu.Lock()
	tail = MAXMSG - 2
	head = MAXMSG - 2
	mu.Unlock()

	post(t, "near-end")
	post(t, "wraps")
	post(t, "around")

	h, tl := peekMsgs(t)
	// Absolute: head=MAXMSG-2, tail=MAXMSG+1
	if h != MAXMSG-2 || tl != MAXMSG+1 {
		t.Fatalf("want head=%d tail=%d, got head=%d tail=%d", MAXMSG-2, MAXMSG+1, h, tl)
	}

	msgs := fetchMsgs(t, h, tl)
	want := []string{"near-end", "wraps", "around"}
	if len(msgs) != len(want) {
		t.Fatalf("want %d messages, got %d: %v", len(want), len(msgs), msgs)
	}
	for i, m := range want {
		if msgs[i] != m {
			t.Fatalf("msg[%d]: want %q got %q", i, m, msgs[i])
		}
	}
}
