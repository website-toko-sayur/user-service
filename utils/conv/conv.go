package conv

import (
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	start := time.Now()

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	log.Debug().
		Dur("duration", time.Since(start)).
		Msg("bcrypt compare")

	return err == nil
}

func LatLngToString(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func StringToInt64(s string) (int64, error) {
	newData, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return newData, nil
}
