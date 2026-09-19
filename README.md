# Pengelola Level DIY Overcooked 2

Platform distribusi level komunitas berbasis paket ekspor [Overcooked2-LevelEditor](https://github.com/gua248/Overcooked2-LevelEditor).

## Fitur

- Penulis mengunggah zip hasil ekspor Level Editor, lalu sistem mem-parsing, memvalidasi, serta mengekstrak tangkapan layar dan metadata secara berurutan di latar belakang
- Penyimpanan privat Tencent Cloud COS, pra-penandatanganan unduhan paket berlaku 6 jam dan gambar 48 jam (diperbarui otomatis secara berkala)
- Halaman beranda publik untuk menjelajah dan mengunduh set level
- Tiga peran hak akses: Penulis / Admin / Super Admin
- Tema siang dan malam dengan tampilan UI dapur yang ramah

## Port

| Lingkungan | Port | Keterangan |
|------------|------|------------|
| API Backend | 14556 | API pengembangan |
| Frontend Vite | 14557 | Frontend pengembangan (proxy `/api` → 14556) |
| Produksi | 14558 | Frontend statis + API dalam satu proses |

## Panduan Cepat

```bash
# 1. Pasang dependensi
make deps

# 2. Salin dan sesuaikan konfigurasi (masukkan kredensial COS)
cp configs/config.example.yaml configs/config.yaml

# 3. Mode pengembangan (jalankan di dua terminal)
make dev-api   # :14556
make dev-fe    # :14557

# 4. Build produksi
make prod      # :14558
```

Akun super admin bawaan: `admin` / `admin123` (bisa diubah di config, hanya dibuat saat aplikasi pertama kali dijalankan)

## Penjelasan Konfigurasi

Lihat [`configs/config.example.yaml`](configs/config.example.yaml)

- **Jika COS dikosongkan**: Menggunakan penyimpanan tiruan lokal di `data/cos-local/` (untuk pengembangan)
- **parser**: Perlu memasang `UnityPy` dan `Pillow` (`pip install -r tools/requirements.txt`)

## Format Unggahan

Format nama zip: `{slug}_v{version}_{yyyyMMdd}.zip`

```
levels/<slug>/info_<slug>
levels/<slug>/s_*
commonW1 / commonW2 / runtime (opsional)
OC2LevelRuntimeLoader.dll (opsional)
```

## Dokumentasi API

- Pengembangan: http://localhost:14556/swagger/index.html
- Swagger JSON: http://localhost:14556/swagger/doc.json

## Tumpukan Teknologi

- Backend: Go + chi + SQLite (WAL) + JWT
- Frontend: React + Vite + Tailwind + framer-motion
- Parser: Subproses Python UnityPy

