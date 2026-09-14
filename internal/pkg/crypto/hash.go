package crypto

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory     uint32 = 32 * 1024
	argon2Time       uint32 = 1
	argon2Threads    uint8  = 2
	argon2SaltLength uint32 = 16
	argon2KeyLength  uint32 = 32
)

var ErrInvalidEncodedPasswordStringLength = errors.New("invalid encoded password string length")

type Argon2Config struct {
	Memory     uint32
	Time       uint32
	Threads    uint8
	SaltLength uint32
	KeyLength  uint32
}

type Argon2Manager struct {
	config *Argon2Config
}

func newArgon2Config() *Argon2Config {
	return &Argon2Config{
		Memory:     argon2Memory,
		Time:       argon2Time,
		Threads:    argon2Threads,
		SaltLength: argon2SaltLength,
		KeyLength:  argon2KeyLength,
	}
}

func NewArgon2() *Argon2Manager {
	config := newArgon2Config()
	return &Argon2Manager{
		config: config,
	}
}

func (a *Argon2Manager) Hash(ctx context.Context, plainPassword string) (string, error) {
	salt := make([]byte, a.config.SaltLength)

	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(plainPassword),
		salt,
		a.config.Time,
		a.config.Memory,
		a.config.Threads,
		a.config.KeyLength,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		a.config.Memory,
		a.config.Time,
		a.config.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func (a *Argon2Manager) Verify(encoded, providedPassword string) (bool, error) {
	memory, time, threads, salt, actualHash, err := parseArgon2EncodedString(encoded)
	if err != nil {
		return false, err
	}

	providedHash := argon2.IDKey(
		[]byte(providedPassword),
		salt,
		time,
		memory,
		threads,
		a.config.KeyLength,
	)

	return subtle.ConstantTimeCompare(actualHash, providedHash) == 1, nil
}

func parseArgon2EncodedString(encoded string) (memory uint32, time uint32, threads uint8, salt []byte, hash []byte, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return 0, 0, 0, nil, nil, ErrInvalidEncodedPasswordStringLength
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return 0, 0, 0, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, err
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, err
	}

	return memory, time, threads, salt, hash, nil
}
