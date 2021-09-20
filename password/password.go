package password

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/scrypt"
	"strings"
)

type Hash struct {
	Hash []byte
	Salt []byte
}

// Create a hash of a password and salt
func createPasswordHash(password string, salt []byte) ([]byte, error) {
	password = strings.TrimSpace(password)

	hash, err := scrypt.Key([]byte(password), salt, 16384, 8, 1, 64)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func NewPasswordHash(password string) (*Hash, error) {
	salt := generateSalt()
	hash, err := createPasswordHash(password, salt)
	if err != nil {
		return nil, err
	}

	return &Hash{Hash: hash, Salt: salt}, nil
}

// Generate a random salt of length 32
func generateSalt() []byte {
	salt := make([]byte, 32)

	_, _ = rand.Read(salt)
	return salt
}

// Value get value of Jsonb
func (h Hash) Value() (driver.Value, error) {
	j, err := json.Marshal(h)
	return j, err
}

// Scan scan value into Hash
func (h *Hash) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, h)
}
