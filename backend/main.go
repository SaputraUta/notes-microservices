package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Note struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

var db *sql.DB

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	if err := initDB(); err != nil {
		log.Fatalf("init db: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /api/notes", listNotes)
	mux.HandleFunc("POST /api/notes", createNote)
	mux.HandleFunc("DELETE /api/notes/{id}", deleteNote)

	log.Println("backend listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// initDB waits for Postgres to be reachable (it may start after us),
// then creates the notes table if it does not exist.
func initDB() error {
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			_, err := db.Exec(`CREATE TABLE IF NOT EXISTS notes (
				id   SERIAL PRIMARY KEY,
				text TEXT NOT NULL
			)`)
			return err
		}
		log.Println("waiting for database...")
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("database not reachable after retries")
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func listNotes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, text FROM notes ORDER BY id DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Text); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		notes = append(notes, n)
	}
	writeJSON(w, notes)
}

func createNote(w http.ResponseWriter, r *http.Request) {
	var n Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if n.Text == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}
	if err := db.QueryRow("INSERT INTO notes (text) VALUES ($1) RETURNING id", n.Text).Scan(&n.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, n)
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := db.Exec("DELETE FROM notes WHERE id = $1", id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
