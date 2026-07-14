package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 🔍 Analogi besar: access + refresh token itu seperti TIKET HARIAN + KARTU MEMBER.
//   - Access token = tiket harian yang cepat hangus (15 menit). Ditunjukkan tiap masuk wahana.
//     Kalau tercecer & dipungut orang, cuma berlaku sebentar -> kerugian kecil.
//   - Refresh token = kartu member tahan lama (7 hari), disimpan baik-baik di dompet. Fungsinya
//     HANYA menukar tiket harian baru saat yang lama hangus — jadi kamu tak perlu daftar ulang (login).
// Prinsipnya: yang sering dipamerkan dibuat cepat kadaluarsa; yang tahan lama disembunyikan & bisa dicabut.

// Pola ACCESS + REFRESH token:
//   - Access token  : umur PENDEK (mis. 15 menit). Dipakai di tiap request.
//   - Refresh token : umur PANJANG (mis. 7 hari). Dipakai HANYA untuk minta
//     access token baru saat yang lama kadaluarsa -> user tak perlu login ulang.
//
// Keuntungan: kalau access token bocor, kadaluarsa cepat. Refresh token disimpan
// lebih aman (HttpOnly cookie) & bisa dicabut (revoke).

var jwtSecret = []byte("ganti-di-produksi-pakai-env")

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
)

type Claims struct {
	UserID uint   `json:"uid"`
	Type   string `json:"typ"` // "access" atau "refresh"
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken = errors.New("token tidak valid")
	ErrWrongType    = errors.New("jenis token salah")
)

func generate(userID uint, typ string, ttl time.Duration) (string, error) {
	c := Claims{
		UserID: userID,
		Type:   typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(jwtSecret)
}

// GenerateTokenPair membuat pasangan access + refresh (dipanggil saat login).
func GenerateTokenPair(userID uint) (access, refresh string, err error) {
	if access, err = generate(userID, "access", accessTTL); err != nil {
		return "", "", err
	}
	if refresh, err = generate(userID, "refresh", refreshTTL); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func parse(tokenStr string) (*Claims, error) {
	var c Claims
	tok, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return jwtSecret, nil
	})
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return &c, nil
}

// RefreshAccessToken menukar refresh token yang valid dengan access token baru.
// Menolak bila yang diberikan bukan refresh token.
func RefreshAccessToken(refreshToken string) (string, error) {
	c, err := parse(refreshToken)
	if err != nil {
		return "", err
	}
	if c.Type != "refresh" {
		return "", ErrWrongType // cegah access token dipakai untuk refresh
	}
	return generate(c.UserID, "access", accessTTL)
}
