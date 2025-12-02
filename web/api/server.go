package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pg-sec-lab/pkg/checker"

	"github.com/jackc/pgx/v5"
	"github.com/rs/cors"
)

type AnalyzeRequest struct {
	DSN string `json:"dsn"`
}

type AnalyzeResponse struct {
	Report *checker.Report `json:"report,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type UploadResponse struct {
	Report *checker.Report `json:"report,omitempty"`
	Error  string          `json:"error,omitempty"`
}

const uploadsDir = "./uploads"

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/api/analyze", handleAnalyze)
	mux.HandleFunc("/api/upload", handleUpload)
	mux.HandleFunc("/api/health", handleHealth)

	// Setup CORS
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 PG SecureLab API server starting on http://localhost:%s", port)
	log.Printf("📊 Endpoints:")
	log.Printf("   POST /api/analyze - Analyze PostgreSQL database")
	log.Printf("   POST /api/upload  - Upload existing report JSON")
	log.Printf("   GET  /api/health  - Health check")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "PG SecureLab API",
		"version": "1.0.0",
	})
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, AnalyzeResponse{
			Error: "Invalid request body",
		})
		return
	}

	if req.DSN == "" {
		respondJSON(w, http.StatusBadRequest, AnalyzeResponse{
			Error: "DSN is required",
		})
		return
	}

	log.Printf("Analyzing database: %s", maskDSN(req.DSN))

	ctx := r.Context()
	conn, err := pgx.Connect(ctx, req.DSN)
	if err != nil {
		log.Printf("Connection failed: %v", err)
		respondJSON(w, http.StatusBadRequest, AnalyzeResponse{
			Error: fmt.Sprintf("Failed to connect: %v", err),
		})
		return
	}
	defer conn.Close(ctx)

	report, err := checker.Analyze(ctx, conn)
	if err != nil {
		log.Printf("Analysis failed: %v", err)
		respondJSON(w, http.StatusInternalServerError, AnalyzeResponse{
			Error: fmt.Sprintf("Analysis failed: %v", err),
		})
		return
	}

	log.Printf("Analysis completed successfully")
	respondJSON(w, http.StatusOK, AnalyzeResponse{
		Report: report,
	})
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondJSON(w, http.StatusBadRequest, UploadResponse{
			Error: "Failed to parse form",
		})
		return
	}

	file, header, err := r.FormFile("report")
	if err != nil {
		respondJSON(w, http.StatusBadRequest, UploadResponse{
			Error: "No file uploaded",
		})
		return
	}
	defer file.Close()

	log.Printf("Uploading file: %s", header.Filename)

	if filepath.Ext(header.Filename) != ".json" {
		respondJSON(w, http.StatusBadRequest, UploadResponse{
			Error: "Only JSON files are allowed",
		})
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, UploadResponse{
			Error: "Failed to read file",
		})
		return
	}

	var report checker.Report
	if err := json.Unmarshal(data, &report); err != nil {
		respondJSON(w, http.StatusBadRequest, UploadResponse{
			Error: fmt.Sprintf("Invalid JSON format: %v", err),
		})
		return
	}

	// Save file to uploads directory
	savePath := filepath.Join(uploadsDir, header.Filename)
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		log.Printf("Warning: Failed to save file: %v", err)
	}

	log.Printf("Report uploaded successfully")
	respondJSON(w, http.StatusOK, UploadResponse{
		Report: &report,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func maskDSN(dsn string) string {
	if idx := strings.Index(dsn, "@"); idx != -1 {
		beforeAt := dsn[:idx]
		if colonIdx := strings.Index(beforeAt, "://"); colonIdx != -1 {
			protocol := beforeAt[:colonIdx+3]
			afterProtocol := beforeAt[colonIdx+3:]
			if pwdIdx := strings.Index(afterProtocol, ":"); pwdIdx != -1 {
				user := afterProtocol[:pwdIdx]
				return protocol + user + ":***" + dsn[idx:]
			}
		}
	}
	return dsn
}
