package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)


type TemplateMetadata struct {
	ID           int      `json:"id,omitempty"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies,omitempty"`
	GitURL       string   `json:"git_url"`
	CreatedAt    string   `json:"created_at,omitempty"`
	License      string   `json:"license,omitempty"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./templates.db")
	if err != nil {
		log.Fatalf("Error opening SQLite database: %v", err)
	}


	createTableQuery := `
	CREATE TABLE IF NOT EXISTS templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		description TEXT,
		dependencies TEXT,
		git_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating templates table: %v", err)
	}
}

func pushTemplateHandler(w http.ResponseWriter, r *http.Request) {
	var metadata TemplateMetadata
	err := json.NewDecoder(r.Body).Decode(&metadata)

	dependenciesJSON, _ := json.Marshal(metadata.Dependencies)
	_, err = db.Exec(`INSERT INTO templates (name, version, description, dependencies, git_url) 
					  VALUES (?, ?, ?, ?, ?)`,
		metadata.Name, metadata.Version, metadata.Description, string(dependenciesJSON), metadata.GitURL)
	if err != nil {
		http.Error(w, "Unable to save template to database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Template %s uploaded successfully!", metadata.Name)
}

func listTemplatesHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT id, name, version, description, dependencies, git_url, created_at FROM templates")
	if err != nil {
		http.Error(w, "Unable to query templates", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var templateList []TemplateMetadata
	for rows.Next() {
		var metadata TemplateMetadata
		var dependenciesJSON string
		if err := rows.Scan(&metadata.ID, &metadata.Name, &metadata.Version, &metadata.Description, &dependenciesJSON, &metadata.GitURL, &metadata.CreatedAt); err != nil {
			http.Error(w, "Unable to read template data", http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal([]byte(dependenciesJSON), &metadata.Dependencies); err != nil {
			http.Error(w, "Unable to parse dependencies", http.StatusInternalServerError)
			return
		}

		templateList = append(templateList, metadata)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templateList)
}

func fetchTemplateHandler(w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Template name missing", http.StatusBadRequest)
		return
	}

	var metadata TemplateMetadata
	var dependenciesJSON string
	err := db.QueryRow("SELECT id, name, version, description, dependencies, git_url, created_at FROM templates WHERE name = ?", name).
		Scan(&metadata.ID, &metadata.Name, &metadata.Version, &metadata.Description, &dependenciesJSON, &metadata.GitURL, &metadata.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Unable to query template", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal([]byte(dependenciesJSON), &metadata.Dependencies); err != nil {
		http.Error(w, "Unable to parse dependencies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

func main() {

	initDB()

	http.HandleFunc("/push-template", pushTemplateHandler)
	http.HandleFunc("/list-templates", listTemplatesHandler)
	http.HandleFunc("/get-template", fetchTemplateHandler)

	log.Println("Template server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
