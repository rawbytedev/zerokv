package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rawbytedev/zerokv"
	"github.com/rawbytedev/zerokv/badgerdb"
	"github.com/rawbytedev/zerokv/memdb"
	"github.com/rawbytedev/zerokv/pebbledb"
)

type entry struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updatedAt"`
}

type store struct {
	mu      sync.RWMutex
	db      zerokv.Core
	name    string
	keys    map[string]entry
	tempDir string
}

type server struct {
	mu      sync.Mutex
	stores  map[string]*store
	current string
}

type response struct {
	OK      bool        `json:"ok"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Backend string      `json:"backend"`
}

func newStore(name string) (*store, error) {
	s := &store{name: name, keys: make(map[string]entry)}
	var err error
	switch name {
	case "memdb":
		s.db, err = memdb.NewMemDataBase(memdb.Config{})
	case "badgerdb":
		s.tempDir, err = os.MkdirTemp("", "zerokv-badger-*")
		if err != nil {
			return nil, err
		}
		s.db, err = badgerdb.NewBadgerDB(badgerdb.Config{Dir: s.tempDir})
	case "pebbledb":
		s.tempDir, err = os.MkdirTemp("", "zerokv-pebble-*")
		if err != nil {
			return nil, err
		}
		s.db, err = pebbledb.NewPebbleDB(pebbledb.Config{Dir: s.tempDir})
	default:
		return nil, fmt.Errorf("unknown backend: %s", name)
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *store) put(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.db.Put(ctx, []byte(key), []byte(value)); err != nil {
		return err
	}
	s.mu.Lock()
	s.keys[key] = entry{Key: key, Value: value, UpdatedAt: time.Now().Format("15:04:05")}
	s.mu.Unlock()
	return nil
}

func (s *store) get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	val, err := s.db.Get(ctx, []byte(key))
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func (s *store) delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.db.Delete(ctx, []byte(key)); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.keys, key)
	s.mu.Unlock()
	return nil
}

func (s *store) scan(prefix string) []entry {
	if s.name == "memdb" {
		s.mu.RLock()
		defer s.mu.RUnlock()
		var result []entry
		for k, v := range s.keys {
			if strings.HasPrefix(k, prefix) {
				result = append(result, v)
			}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
		return result
	}
	it := s.db.Scan([]byte(prefix))
	if it == nil {
		return nil
	}
	defer it.Release()
	var result []entry
	for it.Next() {
		key := string(it.Key())
		val := string(it.Value())
		result = append(result, entry{Key: key, Value: val, UpdatedAt: ""})
	}
	return result
}

func (s *store) allEntries() []entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]entry, 0, len(s.keys))
	for _, v := range s.keys {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result
}

func (s *store) batchOp(ops []map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	batch := s.db.Batch()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, op := range ops {
		switch op["type"] {
		case "put":
			if err := batch.Put([]byte(op["key"]), []byte(op["value"])); err != nil {
				return err
			}
			s.keys[op["key"]] = entry{Key: op["key"], Value: op["value"], UpdatedAt: time.Now().Format("15:04:05")}
		case "delete":
			if err := batch.Delete([]byte(op["key"])); err != nil {
				return err
			}
			delete(s.keys, op["key"])
		}
	}
	return batch.Commit(ctx)
}

func (srv *server) currentStore() *store {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.stores[srv.current]
}

func (srv *server) json(w http.ResponseWriter, status int, r response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(r)
}

func (srv *server) handleSwitch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Backend string `json:"backend"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	srv.mu.Lock()
	if _, ok := srv.stores[body.Backend]; !ok {
		srv.mu.Unlock()
		srv.json(w, 400, response{OK: false, Error: "unknown backend: " + body.Backend})
		return
	}
	srv.current = body.Backend
	srv.mu.Unlock()
	srv.json(w, 200, response{OK: true, Backend: body.Backend, Data: "switched to " + body.Backend})
}

func (srv *server) handlePut(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.Key == "" {
		srv.json(w, 400, response{OK: false, Error: "key is required"})
		return
	}
	s := srv.currentStore()
	if err := s.put(body.Key, body.Value); err != nil {
		srv.json(w, 500, response{OK: false, Error: err.Error(), Backend: s.name})
		return
	}
	srv.json(w, 200, response{OK: true, Backend: s.name, Data: map[string]string{"key": body.Key, "value": body.Value}})
}

func (srv *server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		srv.json(w, 400, response{OK: false, Error: "key is required"})
		return
	}
	s := srv.currentStore()
	val, err := s.get(key)
	if err != nil {
		srv.json(w, 404, response{OK: false, Error: "key not found: " + key, Backend: s.name})
		return
	}
	srv.json(w, 200, response{OK: true, Backend: s.name, Data: map[string]string{"key": key, "value": val}})
}

func (srv *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		srv.json(w, 400, response{OK: false, Error: "key is required"})
		return
	}
	s := srv.currentStore()
	if err := s.delete(key); err != nil {
		srv.json(w, 500, response{OK: false, Error: err.Error(), Backend: s.name})
		return
	}
	srv.json(w, 200, response{OK: true, Backend: s.name, Data: map[string]string{"deleted": key}})
}

func (srv *server) handleScan(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	s := srv.currentStore()
	results := s.scan(prefix)
	srv.json(w, 200, response{OK: true, Backend: s.name, Data: results})
}

func (srv *server) handleKeys(w http.ResponseWriter, r *http.Request) {
	s := srv.currentStore()
	entries := s.allEntries()
	srv.json(w, 200, response{OK: true, Backend: s.name, Data: entries})
}

func (srv *server) handleBatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Ops []map[string]string `json:"ops"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	s := srv.currentStore()
	if err := s.batchOp(body.Ops); err != nil {
		srv.json(w, 500, response{OK: false, Error: err.Error(), Backend: s.name})
		return
	}
	srv.json(w, 200, response{OK: true, Backend: s.name, Data: map[string]int{"ops": len(body.Ops)}})
}

func (srv *server) handleInfo(w http.ResponseWriter, r *http.Request) {
	srv.mu.Lock()
	current := srv.current
	backends := make([]string, 0, len(srv.stores))
	for k := range srv.stores {
		backends = append(backends, k)
	}
	sort.Strings(backends)
	srv.mu.Unlock()
	srv.json(w, 200, response{
		OK:      true,
		Backend: current,
		Data: map[string]interface{}{
			"current":  current,
			"backends": backends,
		},
	})
}

func seedData(s *store) {
	items := []entry{
		{"user:alice", "Engineer @ Acme", ""},
		{"user:bob", "Designer @ Widget Co", ""},
		{"user:carol", "PM @ Startup Inc", ""},
		{"session:abc123", `{"user":"alice","exp":3600}`, ""},
		{"session:xyz789", `{"user":"bob","exp":7200}`, ""},
		{"config:theme", "dark", ""},
		{"config:lang", "en-US", ""},
		{"metric:requests", "42817", ""},
		{"metric:errors", "3", ""},
	}
	for _, it := range items {
		s.put(it.Key, it.Value)
	}
}

func main() {
	backends := []string{"memdb", "badgerdb", "pebbledb"}
	srv := &server{
		stores:  make(map[string]*store),
		current: "memdb",
	}

	log.Println("Initializing backends...")
	for _, name := range backends {
		s, err := newStore(name)
		if err != nil {
			log.Fatalf("failed to init %s: %v", name, err)
		}
		seedData(s)
		srv.stores[name] = s
		log.Printf("  ✓ %s ready", name)
	}

	mux := http.NewServeMux()

	staticDir := filepath.Join(filepath.Dir(os.Args[0]), "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = "cmd/demo/static"
	}
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	mux.HandleFunc("/api/info", srv.handleInfo)
	mux.HandleFunc("/api/switch", srv.handleSwitch)
	mux.HandleFunc("/api/put", srv.handlePut)
	mux.HandleFunc("/api/get", srv.handleGet)
	mux.HandleFunc("/api/delete", srv.handleDelete)
	mux.HandleFunc("/api/scan", srv.handleScan)
	mux.HandleFunc("/api/keys", srv.handleKeys)
	mux.HandleFunc("/api/batch", srv.handleBatch)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet || r.Method == http.MethodOptions {
		} else {
			body, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(strings.NewReader(string(body)))
		}
		mux.ServeHTTP(w, r)
	}

	addr := "0.0.0.0:5000"
	log.Printf("ZeroKV demo running at http://%s", addr)
	if err := http.ListenAndServe(addr, http.HandlerFunc(handler)); err != nil {
		log.Fatal(err)
	}
}
