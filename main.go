package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

type Pelanggan struct {
	ID     int    `json:"id"`
	Nama   string `json:"nama"`
	Alamat string `json:"alamat"`
}

type DataTableResponse struct {
	Data     []Pelanggan `json:"data"`
	Total    int         `json:"total"`
	Filtered int         `json:"filtered"`
	Page     int         `json:"page"`
	Limit    int         `json:"limit"`
}

var (
	db *sql.DB
	mu sync.Mutex
)

func main() {
	var err error
	db, err = sql.Open("sqlite", "./data.db")
	if err != nil {
		log.Fatal(err)
	}

	initDB()

	// Static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/alpine.min.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "alpine.min.js")
	})

	// API endpoints
	http.HandleFunc("/api/data", dataHandler)
	http.HandleFunc("/api/add", addHandler)
	http.HandleFunc("/api/delete", deleteHandler)

	log.Println("Server jalan di :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	db.Exec(`PRAGMA journal_mode=WAL;`)
	db.Exec(`CREATE TABLE IF NOT EXISTS pelanggan (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nama TEXT,
		alamat TEXT
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_nama ON pelanggan(nama)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_alamat ON pelanggan(alamat)`)

	// Check if data exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM pelanggan").Scan(&count)
	if count == 0 {
		seedData()
	}
}

func seedData() {
	log.Println("Generating 1000 dummy data...")

	namaDepan := []string{"Ahmad", "Budi", "Citra", "Dewi", "Eko", "Fitri", "Gunawan", "Hana", "Irfan", "Joko", "Kartika", "Lukman", "Maya", "Nadia", "Omar", "Putri", "Qori", "Rahmat", "Siti", "Taufik", "Umi", "Vina", "Wahyu", "Xena", "Yusuf", "Zahra", "Andika", "Bella", "Chandra", "Diana"}
	namaBelakang := []string{"Pratama", "Wijaya", "Santoso", "Kusuma", "Hidayat", "Rahman", "Saputra", "Putra", "Lestari", "Wati", "Permana", "Sutanto", "Hartono", "Susanto", "Nugroho", "Setiawan", "Kurniawan", "Utama", "Maulana", "Hakim"}
	kota := []string{"Jakarta", "Surabaya", "Bandung", "Medan", "Semarang", "Makassar", "Palembang", "Tangerang", "Depok", "Bekasi", "Malang", "Yogyakarta", "Solo", "Denpasar", "Bogor"}
	jalan := []string{"Jl. Merdeka", "Jl. Sudirman", "Jl. Gatot Subroto", "Jl. Ahmad Yani", "Jl. Diponegoro", "Jl. Pahlawan", "Jl. Kartini", "Jl. Veteran", "Jl. Asia Afrika", "Jl. Imam Bonjol"}

	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("INSERT INTO pelanggan(nama, alamat) VALUES(?, ?)")

	for i := 0; i < 1000; i++ {
		nama := namaDepan[rand.Intn(len(namaDepan))] + " " + namaBelakang[rand.Intn(len(namaBelakang))]
		alamat := jalan[rand.Intn(len(jalan))] + " No. " + strconv.Itoa(rand.Intn(200)+1) + ", " + kota[rand.Intn(len(kota))]
		stmt.Exec(nama, alamat)
	}

	tx.Commit()
	log.Println("1000 dummy data created!")
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Parse query params
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}
	sortCol := r.URL.Query().Get("sort")
	if sortCol == "" {
		sortCol = "id"
	}
	sortOrder := r.URL.Query().Get("order")
	if sortOrder != "desc" {
		sortOrder = "asc"
	}
	searchNama := r.URL.Query().Get("search_nama")
	searchAlamat := r.URL.Query().Get("search_alamat")

	// Validate sort column
	validCols := map[string]bool{"id": true, "nama": true, "alamat": true}
	if !validCols[sortCol] {
		sortCol = "id"
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}

	if searchNama != "" {
		conditions = append(conditions, "nama LIKE ?")
		args = append(args, "%"+searchNama+"%")
	}
	if searchAlamat != "" {
		conditions = append(conditions, "alamat LIKE ?")
		args = append(args, "%"+searchAlamat+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	var total int
	db.QueryRow("SELECT COUNT(*) FROM pelanggan").Scan(&total)

	// Get filtered count
	var filtered int
	countQuery := "SELECT COUNT(*) FROM pelanggan " + whereClause
	db.QueryRow(countQuery, args...).Scan(&filtered)

	// Get data with pagination
	offset := (page - 1) * limit
	query := fmt.Sprintf("SELECT id, nama, alamat FROM pelanggan %s ORDER BY %s %s LIMIT ? OFFSET ?",
		whereClause, sortCol, sortOrder)

	queryArgs := append(args, limit, offset)
	rows, err := db.Query(query, queryArgs...)
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "database error", 500)
		return
	}
	defer rows.Close()

	data := []Pelanggan{}
	for rows.Next() {
		var p Pelanggan
		rows.Scan(&p.ID, &p.Nama, &p.Alamat)
		data = append(data, p)
	}

	response := DataTableResponse{
		Data:     data,
		Total:    total,
		Filtered: filtered,
		Page:     page,
		Limit:    limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	nama := r.URL.Query().Get("nama")
	alamat := r.URL.Query().Get("alamat")
	if nama == "" || alamat == "" {
		http.Error(w, "nama dan alamat wajib diisi", 400)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	res, err := db.Exec("INSERT INTO pelanggan(nama, alamat) VALUES(?, ?)", nama, alamat)
	if err != nil {
		http.Error(w, "database error", 500)
		return
	}
	id, _ := res.LastInsertId()

	pelanggan := Pelanggan{ID: int(id), Nama: nama, Alamat: alamat}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pelanggan)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	mu.Lock()
	defer mu.Unlock()

	db.Exec("DELETE FROM pelanggan WHERE id=?", id)
	w.WriteHeader(204)
}
