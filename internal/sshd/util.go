package sshd

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"os/user"

	"golang.org/x/crypto/ssh"
)

func LoadHostSigner(path string) (ssh.Signer, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(key)
}

func NewMockHostSigner() (ssh.Signer, error) {
	seed, err := deriveDevSeed()
	if err != nil {
		return nil, err
	}

	priv := ed25519.NewKeyFromSeed(seed[:])
	return ssh.NewSignerFromKey(priv)
}

func deriveDevSeed() ([32]byte, error) {
	var out [32]byte

	host, _ := os.Hostname()
	usr, _ := user.Current()

	input := fmt.Sprintf(
		"podsh-dev|%s|%s|%d",
		host,
		usr.Username,
		os.Getuid(),
	)

	sum := sha256.Sum256([]byte(input))

	for i := range 4 {
		binary.BigEndian.PutUint64(
			sum[i*8:(i+1)*8],
			binary.BigEndian.Uint64(sum[i*8:(i+1)*8])^0xA5A5A5A5A5A5A5A5,
		)
	}

	copy(out[:], sum[:32])
	return out, nil
}
