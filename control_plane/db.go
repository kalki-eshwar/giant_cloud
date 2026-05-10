package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := "postgresql://user:pass@localhost:5432/cp?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		// Log but don't strictly fail yet if the local db isn't up during development tests without Docker
		log.Printf("Warning: Database ping failed (is Postgres running?): %v", err)
	}

	DB = db

	createSchema()
}

func createSchema() {
	schema := `
	CREATE TABLE IF NOT EXISTS nodes (
		id SERIAL PRIMARY KEY,
		address VARCHAR(255) UNIQUE NOT NULL,
		capacity INT NOT NULL,
		status VARCHAR(50) DEFAULT 'active'
	);

	CREATE TABLE IF NOT EXISTS manifests (
		id SERIAL PRIMARY KEY,
		file_name VARCHAR(255) NOT NULL,
		total_chunks INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS chunks (
		id SERIAL PRIMARY KEY,
		manifest_id INT REFERENCES manifests(id),
		chunk_index INT NOT NULL,
		node_address VARCHAR(255) NOT NULL,
		status VARCHAR(50) DEFAULT 'allocated'
	);
	`
	_, err := DB.Exec(schema)
	if err != nil {
		log.Printf("Warning: Failed to execute schema setup: %v", err)
	}
}

func LoadNodes(ws *workerServer) {
	if DB == nil {
		return
	}
	rows, err := DB.Query("SELECT address FROM nodes")
	if err != nil {
		log.Printf("Warning: Failed to load nodes from DB: %v", err)
		return
	}
	defer rows.Close()

	ws.mu.Lock()
	defer ws.mu.Unlock()
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err == nil {
			found := false
			for _, n := range ws.nodes {
				if n == address {
					found = true
					break
				}
			}
			if !found {
				ws.nodes = append(ws.nodes, address)
			}
		}
	}
}
