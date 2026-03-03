package cache

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/ssh"
)

func ComputeKey(pubKey []byte) string {
	sum := sha256.Sum256(pubKey)
	return "identity:" + hex.EncodeToString(sum[:])
}

func NormalizePublicKey(pkBytes []byte) ([]byte, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pkBytes)
	if err != nil {
		return nil, err
	}

	return pubKey.Marshal(), nil
}
