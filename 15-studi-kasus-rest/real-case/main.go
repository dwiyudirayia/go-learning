// REAL-CASE Modul 15 (studi kasus REST) — stack produksi lengkap:
// Fiber + PostgreSQL (user) + Redis (refresh token) + JWT + bcrypt.
//
// Versi advanced/ memakai repo in-memory. Versi ini memakai infra nyata dan
// diuji end-to-end via app.Test (tanpa port). Auto-skip bila env kosong.
//
// Jalankan nyata:
//
//	docker compose -f 15-studi-kasus-rest/real-case/docker-compose.yml up -d
//	POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' \
//	REDIS_ADDR=127.0.0.1:6379 go run ./15-studi-kasus-rest/real-case
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("ganti-dengan-secret-dari-env")

// ---------------------------------------------------------------------------
// LAPISAN REPOSITORY (PostgreSQL) — akses data user.
// ---------------------------------------------------------------------------

// UserRepo membungkus pool Postgres untuk operasi tabel users.
type UserRepo struct{ pool *pgxpool.Pool }

// migrate memastikan tabel users ada (produksi: golang-migrate, modul 21).
func (r *UserRepo) migrate(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS users(
		id BIGSERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		hash_pw TEXT NOT NULL)`)
	return err
}

// Create menyimpan user baru. Mengembalikan error bila email sudah ada
// (dilanggar UNIQUE constraint) atau kegagalan DB lain.
func (r *UserRepo) Create(ctx context.Context, email, hashPw string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO users(email, hash_pw) VALUES($1,$2)`, email, hashPw)
	return err
}

// FindHashByEmail mengambil hash password milik email. ok=false bila tak ada.
func (r *UserRepo) FindHashByEmail(ctx context.Context, email string) (hash string, ok bool, err error) {
	err = r.pool.QueryRow(ctx, `SELECT hash_pw FROM users WHERE email=$1`, email).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return hash, true, nil
}

// ---------------------------------------------------------------------------
// LAPISAN TOKEN STORE (Redis) — menyimpan refresh token yang bisa di-revoke.
// ---------------------------------------------------------------------------

// TokenStore menyimpan refresh token di Redis (dengan TTL & bisa dihapus/revoke).
type TokenStore struct{ rdb *redis.Client }

// SaveRefresh menyimpan token -> email dengan masa berlaku (TTL).
func (t *TokenStore) SaveRefresh(ctx context.Context, token, email string, ttl time.Duration) error {
	return t.rdb.Set(ctx, "refresh:"+token, email, ttl).Err()
}

// ---------------------------------------------------------------------------
// LAPISAN SERVICE — aturan bisnis (tak tahu HTTP).
// ---------------------------------------------------------------------------

type AuthService struct {
	users  *UserRepo
	tokens *TokenStore
}

// Register mem-VALIDASI input, meng-HASH password dgn bcrypt, lalu menyimpan.
// bcrypt otomatis bergaram & lambat -> tahan brute force.
func (s *AuthService) Register(ctx context.Context, email, pw string) error {
	if email == "" || len(pw) < 6 {
		return errors.New("email wajib & password minimal 6 karakter")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.Create(ctx, email, string(hash))
}

// Login memverifikasi kredensial lalu menerbitkan access token (JWT) + refresh
// token (disimpan di Redis). Mengembalikan pesan generik agar tak membocorkan
// apakah email atau password yang salah.
func (s *AuthService) Login(ctx context.Context, email, pw string) (accessToken, refreshToken string, err error) {
	hash, ok, err := s.users.FindHashByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	// bcrypt.CompareHashAndPassword = perbandingan tahan timing-attack.
	if !ok || bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) != nil {
		return "", "", errors.New("kredensial salah")
	}
	// Access token JWT berumur pendek.
	accessToken, err = issueJWT(email, 15*time.Minute)
	if err != nil {
		return "", "", err
	}
	// Refresh token acak, disimpan di Redis (bisa di-revoke), berumur panjang.
	refreshToken = randomToken()
	if err := s.tokens.SaveRefresh(ctx, refreshToken, email, 24*time.Hour); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// issueJWT membuat token HS256 berisi klaim subject (email) & waktu kedaluwarsa.
func issueJWT(email string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(ttl).Unix(),
		"iat": time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

// parseJWT memvalidasi tanda tangan + kedaluwarsa, mengembalikan email (sub).
func parseJWT(tokenStr string) (string, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		// Pastikan algoritma sesuai -> cegah serangan pergantian algoritma.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("algoritma tak terduga")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}
	sub, _ := tok.Claims.(jwt.MapClaims)["sub"].(string)
	return sub, nil
}

// randomToken menghasilkan token acak 32 hex char (untuk refresh token).
func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// LAPISAN HANDLER (Fiber) — batas HTTP.
// ---------------------------------------------------------------------------

// buildApp merakit rute Fiber dan menyuntikkan service.
func buildApp(svc *AuthService) *fiber.App {
	app := fiber.New()

	// POST /register -> buat akun.
	app.Post("/register", func(c *fiber.Ctx) error {
		var in struct{ Email, Password string }
		_ = c.BodyParser(&in)
		if err := svc.Register(c.Context(), in.Email, in.Password); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fiber.StatusCreated)
	})

	// POST /login -> kembalikan access + refresh token.
	app.Post("/login", func(c *fiber.Ctx) error {
		var in struct{ Email, Password string }
		_ = c.BodyParser(&in)
		access, refresh, err := svc.Login(c.Context(), in.Email, in.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"access_token": access, "refresh_token": refresh})
	})

	// GET /me -> endpoint terproteksi; butuh header Authorization: Bearer <jwt>.
	app.Get("/me", func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		token := strings.TrimPrefix(auth, "Bearer ")
		email, err := parseJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token tidak valid"})
		}
		return c.JSON(fiber.Map{"email": email})
	})

	return app
}

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	addr := os.Getenv("REDIS_ADDR")
	if dsn == "" || addr == "" {
		fmt.Println("⏭️  DILEWATI: set POSTGRES_DSN & REDIS_ADDR untuk versi nyata.")
		fmt.Println("   docker compose -f 15-studi-kasus-rest/real-case/docker-compose.yml up -d")
		fmt.Println("   POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' \\")
		fmt.Println("   REDIS_ADDR=127.0.0.1:6379 go run ./15-studi-kasus-rest/real-case")
		return
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic("gagal konek Postgres: " + err.Error())
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("gagal konek Redis: " + err.Error())
	}

	// COMPOSITION ROOT: rakit repo + token store -> service -> handler.
	repo := &UserRepo{pool: pool}
	if err := repo.migrate(ctx); err != nil {
		panic(err)
	}
	_, _ = pool.Exec(ctx, `TRUNCATE users RESTART IDENTITY`) // mulai bersih utk demo
	svc := &AuthService{users: repo, tokens: &TokenStore{rdb: rdb}}
	app := buildApp(svc)

	// Uji end-to-end via app.Test (menembus semua lapisan + infra nyata).
	email := "budi@mail.com"
	fmt.Println("== 1. register ==")
	fmt.Println(" ", call(app, "POST", "/register", `{"Email":"`+email+`","Password":"rahasia123"}`, ""))

	fmt.Println("== 2. login -> dapat token ==")
	loginResp := call(app, "POST", "/login", `{"Email":"`+email+`","Password":"rahasia123"}`, "")
	fmt.Println(" ", ringkas(loginResp))
	access := ambilAccessToken(loginResp)

	fmt.Println("== 3. akses /me dengan JWT ==")
	fmt.Println(" ", call(app, "GET", "/me", "", "Bearer "+access))

	fmt.Println("== 4. login salah password -> 401 ==")
	fmt.Println(" ", call(app, "POST", "/login", `{"Email":"`+email+`","Password":"salah"}`, ""))
}

// call mengirim satu request ke app Fiber via app.Test dan mengembalikan
// "status body". Param authHeader diisi bila endpoint butuh Authorization.
func call(app *fiber.App, method, path, body, authHeader string) string {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := app.Test(req)
	if err != nil {
		return "err: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return fmt.Sprintf("%d %s", resp.StatusCode, b)
}

// ambilAccessToken mengekstrak nilai access_token dari body login (parse sederhana).
func ambilAccessToken(loginResp string) string {
	const k = `"access_token":"`
	i := strings.Index(loginResp, k)
	if i < 0 {
		return ""
	}
	rest := loginResp[i+len(k):]
	j := strings.IndexByte(rest, '"')
	if j < 0 {
		return ""
	}
	return rest[:j]
}

// ringkas memotong string panjang agar output rapi.
func ringkas(s string) string {
	if len(s) > 90 {
		return s[:90] + "…\"}"
	}
	return s
}
