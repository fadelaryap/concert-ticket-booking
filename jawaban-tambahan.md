### Jawaban Pertanyaan Teknis Tambahan

#### 1. Optimasi SQL & Indeks

**Struktur Database Relasional:**

Sistem ini memiliki tabel-tabel utama yang saling berelasi untuk mengelola data konser, tiket, booking, pengguna, dan pembayaran. Tabel-tabel tersebut meliputi:

* **`users`**: Menyimpan informasi pengguna (`id`, `username`, `email`, `password`, `role`, dll.).
* **`concerts`**: Menyimpan detail konser (`id`, `nama`, `artis`, `tanggal`, `venue`, `total_seats`, `available_seats`, dll.).
* **`ticket_classes`**: Menyimpan kategori tiket untuk setiap konser (`id`, `concert_id`, `nama kelas`, `harga`, `jumlah kursi per kelas`).
* **`seats`**: Menyimpan detail setiap kursi (`id`, `concert_id`, `ticket_class_id`, `nomor kursi`, `status`, `user_id`, `booking_id`).
* **`bookings`**: Menyimpan informasi booking tiket (`id` UUID, `user_id`, `concert_id`, `total_price`, `status`, `expires_at`, dll.).
* **`booking_seats`**: Tabel join many-to-many antara `bookings` dan `seats` untuk melacak kursi mana yang termasuk dalam booking tertentu.
* **`buyers`**: Menyimpan informasi pembeli untuk setiap booking (`id`, `booking_id`, `nama lengkap`, `nomor telepon`, `email`, `nomor KTP`).
* **`ticket_holders`**: Menyimpan informasi pemegang tiket jika berbeda dari pembeli (`id`, `booking_id`, `nama lengkap`, `nomor KTP`).
* **`payments`**: Menyimpan detail transaksi pembayaran (`id`, `booking_id`, `amount`, `payment_method`, `transaction_id`, `status`).

**Strategi Indeks yang Tepat:**

* **`users`**:
    * `username` dan `email`: Sudah `UNIQUE`, yang otomatis membuat indeks, mempercepat login dan pencarian pengguna.
* **`concerts`**:
    * `date`: Indeks pada kolom ini akan mempercepat pencarian konser berdasarkan tanggal atau untuk event yang akan datang.
    * `status`: Mempercepat filter konser berdasarkan status (misalnya, `active`, `pending_seat_creation`).
    * `name`: Berguna untuk pencarian cepat berdasarkan nama konser.
* **`ticket_classes`**:
    * `concert_id`: Diindeks karena ini adalah Foreign Key (FK), penting untuk mengambil kelas tiket untuk konser spesifik.
    * `concert_id`, `name`: Indeks komposit akan mempercepat pencarian kelas tiket berdasarkan konser dan nama kelas.
* **`seats`**:
    * `concert_id` dan `ticket_class_id`: Keduanya diindeks sebagai FK.
    * `status`: Sangat penting untuk menemukan kursi yang `available`, `reserved`, atau `booked`.
    * `user_id`, `booking_id`: Mempercepat pencarian kursi berdasarkan pengguna atau booking terkait.
    * `concert_id`, `ticket_class_id`, `seat_number`: Indeks unik komposit memastikan kursi unik per kelas per konser dan mempercepat pencarian kursi spesifik.
* **`bookings`**:
    * `user_id`: Sangat penting untuk mengambil semua booking milik seorang pengguna.
    * `concert_id`: Mempercepat pencarian booking untuk konser tertentu.
    * `status`: Krusial untuk memfilter booking berdasarkan status (`pending`, `confirmed`, `cancelled`).
    * `expires_at`: Sangat penting untuk fitur pembatalan booking otomatis yang kedaluwarsa.
    * `id`: Primary key (UUID string) secara otomatis diindeks.
* **`buyers` dan `ticket_holders`**:
    * `booking_id`: Diindeks sebagai FK, mempercepat pengambilan detail pembeli/pemegang tiket untuk booking.
    * `ktp_number`: Sudah `UNIQUE`, otomatis diindeks.
* **`payments`**:
    * `booking_id`: Diindeks untuk mencari pembayaran terkait booking.
    * `transaction_id`: Sudah `UNIQUE`, otomatis diindeks, penting untuk melacak transaksi gateway.
    * `status`: Berguna untuk memfilter pembayaran berdasarkan statusnya.

#### 2. Optimasi Beban Tinggi pada Database

Mengurangi beban pada database (DB) saat endpoint menerima 100.000 hit/detik memerlukan strategi yang berfokus langsung pada efisiensi interaksi database.

1.  **Caching (Redis)**:
    Penyimpanan data yang sering diakses di cache seperti Redis dapat mengurangi jumlah permintaan baca langsung ke database. Data seperti ketersediaan kursi konser atau detail konser yang jarang berubah dapat diambil dari cache yang lebih cepat, sehingga beban pada database MySQL berkurang.

2.  **Message Queues (RabbitMQ)**:
    Penggunaan antrean pesan seperti RabbitMQ memungkinkan operasi yang memakan waktu atau berat untuk dijalankan secara asinkron di latar belakang. Contohnya, pembuatan kursi setelah konser dibuat dapat di-offload ke worker terpisah. Ini membebaskan thread utama layanan dari menunggu operasi database yang panjang, sehingga mengurangi beban sinkron pada DB.

3.  **Database Connection Pooling**:
    Pengelolaan koneksi database yang efisien melalui connection pooling sangat penting. Mengkonfigurasi parameter seperti jumlah koneksi idle dan koneksi terbuka maksimum membantu menggunakan kembali koneksi yang ada, mengurangi overhead pembuatan koneksi baru ke database dan menjaga stabilitas di bawah beban tinggi.

4.  **Read Replicas (Replika Baca)**:
    Untuk aplikasi dengan banyak operasi baca, penggunaan replika database khusus untuk operasi baca dapat mendistribusikan beban. Permintaan baca diarahkan ke replika, sementara database master hanya menangani operasi tulis. Ini secara signifikan meningkatkan throughput untuk operasi baca.

5.  **Horizontal Scaling Database**:
    Meningkatkan jumlah instance database, seperti replika baca atau shard, adalah cara langsung untuk mendistribusikan beban. Setiap instance dapat menangani sebagian dari permintaan, sehingga total beban per instance database berkurang.

6.  **Optimasi Query dan Index**:
    Query SQL yang ditulis secara efisien dan penggunaan indeks yang tepat adalah fundamental. Indeks memungkinkan database menemukan data yang relevan dengan cepat tanpa harus memindai seluruh tabel. Query yang tidak teroptimasi dapat membebani database dengan cepat, bahkan dengan strategi lain.

7.  **Database Sharding/Partitioning**:
    Untuk tabel yang sangat besar, membagi data ke beberapa database atau partisi dapat mendistribusikan beban I/O. Data dipisahkan berdasarkan kriteria tertentu (misalnya, ID konser atau tanggal), sehingga setiap permintaan hanya perlu mengakses subset data yang lebih kecil.

#### 3. Keamanan API

Endpoint `POST /api/users/register` adalah salah satu endpoint paling kritis dalam aplikasi karena ini adalah titik masuk bagi pengguna baru. Mengamankannya sangat penting untuk mencegah penyalahgunaan. Berikut minimal 5 langkah untuk mengamankannya:

1.  **Validasi dan Sanitasi Input**: Memastikan semua data yang diterima (username, email, password) divalidasi dengan ketat terhadap format, panjang, dan tipe yang diharapkan. Sanitasi input untuk mencegah serangan seperti SQL Injection atau Cross-Site Scripting (XSS) jika data tersebut akan ditampilkan kembali.
2.  **Hashing Kata Sandi yang Kuat**: Jangan pernah menyimpan kata sandi dalam bentuk plaintext di database. Gunakan algoritma hashing satu arah yang kuat dan lambat seperti bcrypt, Argon2, atau scrypt. Selalu gunakan salt unik per pengguna untuk mencegah serangan rainbow table.
3.  **Pembatasan Tingkat Permintaan (Rate Limiting)**: Membatasi jumlah permintaan yang dapat dibuat dari satu IP address atau pengguna dalam jangka waktu tertentu. Ini mencegah serangan brute-force untuk mencoba mendaftar akun secara massal atau mencoba menebak kredensial.
4.  **Integrasi CAPTCHA/reCAPTCHA**: Menerapkan mekanisme CAPTCHA (seperti Google reCAPTCHA) pada formulir pendaftaran. Ini membantu membedakan antara pengguna manusia dan bot otomatis, mencegah pendaftaran spam dan serangan bot.
5.  **Verifikasi Email**: Setelah pendaftaran, mengirim email konfirmasi ke alamat email yang diberikan. Akun pengguna harus tetap tidak aktif atau memiliki fungsionalitas terbatas sampai email dikonfirmasi dengan mengklik tautan unik. Ini mencegah pendaftaran dengan email palsu dan menambahkan lapisan verifikasi identitas.
6.  **Penerapan HTTPS/TLS**: Memastikan semua komunikasi ke dan dari endpoint API dienkripsi menggunakan HTTPS/TLS. Ini melindungi data sensitif (seperti kredensial pendaftaran) dari intersepsi oleh pihak ketiga saat transit.
7.  **Konfigurasi CORS (Cross-Origin Resource Sharing)**: Mengkonfigurasi CORS dengan benar untuk hanya mengizinkan permintaan dari domain frontend yang dipercaya. Ini mencegah domain jahat melakukan permintaan lintas-origin yang tidak sah ke API.

#### 4. Alur Kerja CI/CD

Continuous Integration (CI) dan Continuous Deployment (CD) adalah praktik DevOps yang bertujuan untuk mengotomatisasi proses pengembangan perangkat lunak mulai dari integrasi kode hingga deployment ke produksi.

**Fase 1: Continuous Integration (CI)**

Tujuan CI adalah untuk mengintegrasikan perubahan kode dari beberapa developer ke dalam satu repositori utama secara sering, dan memverifikasi perubahan tersebut secara otomatis.

1.  **Developer Melakukan Commit Kode**:
    Developer menulis kode dan melakukan `git commit` lalu `git push` perubahan ke branch (misalnya `feature` atau `develop`) di Git Repository (misalnya GitHub, GitLab).

2.  **Pemicu Server CI**:
    Webhook di repositori Git secara otomatis memberi tahu Server CI (misalnya Jenkins, GitLab CI, GitHub Actions) setiap kali ada `git push` baru.

3.  **Pengambilan Kode**:
    Server CI menerima notifikasi dan menarik (pull) kode terbaru dari repositori Git.

4.  **Tahap Build**:
    * **Backend (Layanan Go)**: Setiap layanan Go akan dikompilasi. Docker images untuk setiap layanan akan dibangun, berisi aplikasi Go yang sudah dikompilasi dan dependensinya.
    * **Frontend (Aplikasi React)**: Aplikasi React akan dibangun, menghasilkan aset statis (HTML, CSS, JavaScript) yang dioptimalkan untuk produksi. Docker image untuk aplikasi frontend juga akan dibangun.

5.  **Tahap Pengujian**:
    * **Unit Tests**: Jalankan semua unit test untuk backend dan frontend.
    * **Integration Tests**: Jalankan tes yang memverifikasi interaksi antar layanan.
    * **Analisis Statis / Linting**: Jalankan alat seperti `go vet`, `golangci-lint` untuk Go, dan ESLint/TypeScript ESLint untuk React untuk memeriksa kualitas kode.

6.  **Penyimpanan Artefak**:
    Jika semua tahap build dan test berhasil, Docker images akan di-tag dengan versi dan diunggah ke Container Registry (misalnya Docker Hub, Google Container Registry). Ini adalah "artefak" yang siap untuk deployment.

**Fase 2: Deployment Berkelanjutan (CD)**

Tujuan CD adalah untuk secara otomatis (atau semi-otomatis dengan persetujuan manual) mendeploy aplikasi yang sudah terverifikasi dari CI ke lingkungan yang berbeda (staging, production).

1.  **Pemicu Deployment**:
    Jika semua tahap CI berhasil, deployment ke lingkungan staging bisa otomatis. Deployment ke produksi seringkali memerlukan persetujuan manual.

2.  **Eksekusi Alat Deployment**:
    Alat CD (misalnya Kubernetes dengan ArgoCD/Flux, Ansible, Terraform) akan dipicu. Alat ini akan mengakses Container Registry untuk menarik (pull) Docker images versi terbaru.

3.  **Penyediaan / Pembaruan Infrastruktur (jika diperlukan)**:
    Jika ada perubahan infrastruktur, alat seperti Terraform atau Ansible akan memperbarui infrastruktur cloud/server.

4.  **Deployment Aplikasi**:
    * **Pembaruan Bergulir (Rolling Updates)**: Layanan akan di-deploy menggunakan strategi pembaruan bergulir untuk meminimalkan downtime. Instans lama diganti secara bertahap dengan instans baru.
    * Pembaruan konfigurasi spesifik lingkungan target.

5.  **Verifikasi Pasca-Deployment / Uji Asap (Smoke Tests)**:
    Setelah deployment, jalankan serangkaian tes cepat ("uji asap") untuk memastikan aplikasi berfungsi dengan benar di lingkungan baru.

6.  **Pemantauan dan Peringatan**:
    Sistem pemantauan akan terus memantau performa aplikasi dan kesehatan layanan. Jika ada anomali atau error, sistem akan memicu peringatan.

7.  **Strategi Rollback**:
    Jika terjadi masalah serius setelah deployment, harus ada kemampuan untuk dengan cepat mengembalikan (rollback) ke versi aplikasi yang stabil sebelumnya.

#### 5. Tinjauan Kode & Debugging (Golang)

Berikut potongan kode handler API menggunakan Gin:

```go
r.GET("/user/:id", func(c *gin.Context) {
    id := c.Param("id")
    user, err := db.FindUserByID(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }
    if user == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    c.JSON(http.StatusOK, user)
})
```

Identifikasi Masalah/Bug dan Penjelasan Perbaikan:

Saya mengidentifikasi beberapa masalah pada kode di atas:

1.  **Masalah 1: Tipe Data Parameter id Tidak Konsisten**

    Kesalahan: c.Param("id") mengembalikan string. Namun, fungsi db.FindUserByID(id) kemungkinan besar mengharapkan tipe data numerik (uint atau int) karena id biasanya adalah primary key numerik di database. Meneruskan string secara langsung akan menyebabkan error kompilasi atau runtime.
    Perbaikan: Lakukan konversi tipe dari string ke uint menggunakan strconv.ParseUint dan tambahkan penanganan error jika konversi gagal (misalnya, ID yang tidak valid).

2.  **Masalah 2: Penanganan Error Database yang Tidak Spesifik dan Ambigu**

    Kesalahan: Blok if err != nil menangkap semua error dari db.FindUserByID, termasuk ketika pengguna tidak ditemukan (gorm.ErrRecordNotFound). Akibatnya, gorm.ErrRecordNotFound akan diperlakukan sebagai "Internal server error" (HTTP 500) alih-alih "Not Found" (HTTP 404). Blok if user == nil mungkin tidak pernah tercapai untuk kasus "not found" karena err sudah ditangkap lebih dulu.
    Perbaikan: Gunakan errors.Is(err, gorm.ErrRecordNotFound) untuk secara spesifik memeriksa error "record not found" dan kembalikan HTTP 404. Untuk error lain, kembalikan HTTP 500.

3.  **Masalah 3: Kurangnya Otorisasi (Kerentanan Keamanan Kritis)**

    Kesalahan: Endpoint ini memungkinkan siapa pun yang dapat mengaksesnya untuk melihat profil pengguna mana pun hanya dengan mengetahui ID mereka. Ini adalah celah keamanan yang serius. Pengguna seharusnya hanya dapat melihat profil mereka sendiri, atau endpoint ini harus dilindungi dengan pemeriksaan peran (misalnya, hanya admin).
    Perbaikan: Dapatkan ID pengguna yang terautentikasi dari konteks Gin (yang seharusnya disetel oleh middleware autentikasi) dan bandingkan dengan id yang diminta di URL. Jika tidak cocok, kembalikan HTTP 403 Forbidden.

4.  **Masalah 4: Pesan Error yang Tidak Informatif dan Kurangnya Logging**

    Kesalahan: Pesan error "Internal server error" terlalu generik. Ini tidak membantu dalam debugging. Error yang sebenarnya (err) tidak dicatat ke log aplikasi.
    Perbaikan: Catat error internal secara detail menggunakan logger aplikasi (misalnya, utils.LogError("Database error: %v", err)) dan berikan pesan error yang lebih umum namun aman kepada klien.

**Versi Kode yang Sudah Diperbaiki:**

```go
import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend/user-service/middlewares"
	"backend/user-service/utils"
)

r.GET("/user/:id", middlewares.AuthMiddleware(), func(c *gin.Context) {
    idParam := c.Param("id")
    requestedID, err := strconv.ParseUint(idParam, 10, 32)
    if err != nil {
        utils.LogWarning("Invalid user ID format in request: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
        return
    }

    authUserID, exists := c.Get("userID")
    if !exists {
        utils.LogError("Authenticated user ID not found in context for /user/:id endpoint")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication context error"})
        return
    }

    if uint(requestedID) != authUserID.(uint) {
        utils.LogWarning("Unauthorized access attempt: User %d tried to access profile of ID %d", authUserID.(uint), requestedID)
        c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only view your own profile"})
        return
    }

    user, err := db.FindUserByID(uint(requestedID))
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            utils.LogInfo("User ID %d not found in DB", requestedID)
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        utils.LogError("Database error fetching user ID %d: %v", requestedID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
        return
    }

    c.JSON(http.StatusOK, user)
})
```
