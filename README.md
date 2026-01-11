# Todo List - Go + Alpine.js (Windows 7 Style)

Aplikasi Todo List sederhana dengan UI bergaya Windows 7. Backend Go, frontend Alpine.js, database SQLite.

![Todo List](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Alpine.js](https://img.shields.io/badge/Alpine.js-8BC0D0?style=flat&logo=alpine.js&logoColor=black)
![SQLite](https://img.shields.io/badge/SQLite-003B57?style=flat&logo=sqlite&logoColor=white)

## Fitur

- ✅ Tambah tugas
- ✅ Hapus tugas
- ✅ Tandai selesai
- ✅ Window draggable (bisa digeser)
- ✅ Tanpa reload halaman

## Instalasi

1. **Clone repository**
   ```bash
   git clone https://github.com/falehjamal/todolist-go-windows-design.git
   cd todolist-go-windows-design
   ```

2. **Install dependency Go**
   ```bash
   go mod tidy
   ```

3. **Jalankan server**
   ```bash
   go run main.go
   ```

4. **Buka browser**
   ```
   http://localhost:8080
   ```

## Struktur File

```
├── main.go          # Backend Go
├── index.html       # Frontend Alpine.js
├── alpine.min.js    # Alpine.js library
├── go.mod           # Go module
└── data.db          # SQLite database (auto-generated)
```

## Lisensi

MIT
