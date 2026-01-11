package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	_ "modernc.org/sqlite"
)

type Todo struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

var (
	db    *sql.DB
	cache = make(map[int]Todo)
	mu    sync.Mutex
)

func main() {
	var err error
	db, err = sql.Open("sqlite", "./data.db")
	if err != nil {
		log.Fatal(err)
	}

	initDB()
	loadCache()

	// Static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/alpine.min.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "alpine.min.js")
	})

	// API endpoints
	http.HandleFunc("/api/todos", todosHandler)
	http.HandleFunc("/api/add", addHandler)
	http.HandleFunc("/api/toggle", toggleHandler)
	http.HandleFunc("/api/delete", deleteHandler)

	log.Println("Server jalan di :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	db.Exec(`PRAGMA journal_mode=WAL;`)
	db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT,
		completed INTEGER DEFAULT 0
	)`)
}

func loadCache() {
	rows, err := db.Query("SELECT id, text, completed FROM todos")
	if err != nil {
		log.Println("loadCache error:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t Todo
		var completed int
		rows.Scan(&t.ID, &t.Text, &completed)
		t.Completed = completed == 1
		cache[t.ID] = t
	}
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	list := make([]Todo, 0, len(cache))
	for _, v := range cache {
		list = append(list, v)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "text kosong", 400)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	res, err := db.Exec("INSERT INTO todos(text, completed) VALUES(?, 0)", text)
	if err != nil {
		log.Println("addHandler db.Exec error:", err)
		http.Error(w, "database error", 500)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Println("addHandler LastInsertId error:", err)
		http.Error(w, "database error", 500)
		return
	}

	todo := Todo{ID: int(id), Text: text, Completed: false}
	cache[todo.ID] = todo

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func toggleHandler(w http.ResponseWriter, r *http.Request) {
	id := atoi(r.URL.Query().Get("id"))

	mu.Lock()
	defer mu.Unlock()

	todo, exists := cache[id]
	if !exists {
		http.Error(w, "not found", 404)
		return
	}

	todo.Completed = !todo.Completed
	cache[id] = todo

	completedInt := 0
	if todo.Completed {
		completedInt = 1
	}
	db.Exec("UPDATE todos SET completed=? WHERE id=?", completedInt, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	mu.Lock()
	defer mu.Unlock()

	db.Exec("DELETE FROM todos WHERE id=?", id)
	delete(cache, atoi(id))

	w.WriteHeader(204)
}

func atoi(s string) int {
	var n int
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
