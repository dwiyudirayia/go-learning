// Studi Kasus REST API — Task Manager dengan Auth JWT (Fiber + GORM).
// Arsitektur berlapis: handler -> service -> repository -> DB.
//
// Jalankan:
//
//	go run ./15-studi-kasus-rest            # :3000
//	PORT=8099 go run ./15-studi-kasus-rest  # port lain
//
// Alur pakai:
//  1. POST /auth/register {name,email,password}
//  2. POST /auth/login    {email,password}  -> {token}
//  3. Panggil /tasks dengan header: Authorization: Bearer <token>
//
// Verifikasi otomatis: go test ./15-studi-kasus-rest
package main

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"go-learning/15-studi-kasus-rest/internal/handler"
	"go-learning/15-studi-kasus-rest/internal/model"
	"go-learning/15-studi-kasus-rest/internal/repository"
	"go-learning/15-studi-kasus-rest/internal/service"
)

// 🔍 Analogi besar: arsitektur berlapis itu seperti RESTORAN dengan pembagian tugas jelas:
//   - handler  = PELAYAN: menerima pesanan (HTTP request), menyampaikan hasil. Tak memasak.
//   - service  = KOKI: logika bisnis (aturan, validasi, hitung). Tak tahu soal HTTP/DB.
//   - repository = PENJAGA GUDANG: satu-satunya yang menyentuh database.
// Kenapa dipisah? Supaya tiap bagian bisa diganti/diuji sendiri. Ganti Fiber ke net/http? cukup
// ubah pelayan. Ganti SQLite ke Postgres? cukup ubah gudang. Koki (aturan bisnis) tak tersentuh.

// 🔍 Analogi: "wiring" di buildApp itu seperti MERANGKAI PIPA dari gudang -> koki -> pelayan
// saat restoran buka. Ini "dependency injection manual": tiap lapisan diberi bahan yang
// dibutuhkannya dari luar. Karena buildApp menerima db, test bisa menyuntik DB in-memory.

// buildApp merangkai seluruh dependency (dependency injection manual) dan
// mengembalikan app Fiber yang siap. Dipakai oleh main dan oleh test.
func buildApp(db *gorm.DB) *fiber.App {
	// Wiring lapisan: repo -> service -> handler.
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	authSvc := service.NewAuthService(userRepo)
	taskSvc := service.NewTaskService(taskRepo)
	h := handler.New(authSvc, taskSvc)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if fe, ok := err.(*fiber.Error); ok {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(recover.New())
	app.Use(logger.New())

	h.Register(app)
	return app
}

// openDB membuka koneksi GORM + migrasi tabel.
func openDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.User{}, &model.Task{}); err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := openDB("taskmanager.db")
	if err != nil {
		log.Fatal(err)
	}

	app := buildApp(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Task Manager API di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
