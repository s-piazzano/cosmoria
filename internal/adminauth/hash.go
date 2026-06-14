package adminauth

import "golang.org/x/crypto/bcrypt"

func HashPassword(plain string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
}

func CheckPassword(plain string, hash []byte) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(plain)) == nil
}
