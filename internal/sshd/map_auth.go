package sshd

import (
	"context"
	"errors"
)

// MapAuthenticator is a simple dev authenticator.
// Maps public key (as bytes) => Identity.
type MapAuthenticator struct {
	keys map[string]*Identity // key = string(pubKey)
}

// NewMapAuthenticator creates a MapAuthenticator from a map
func NewMapAuthenticator(keys map[string]*Identity) *MapAuthenticator {
	return &MapAuthenticator{
		keys: keys,
	}
}

// Authenticate implements PublicKeyAuthenticator
func (a *MapAuthenticator) Authenticate(ctx context.Context, meta ConnMetadata, pubKey []byte) (*Identity, error) {
	id, ok := a.keys[string(pubKey)]
	if !ok {
		return nil, errors.New("unauthorized: unknown public key")
	}

	// optionally attach remote address
	identity := *id
	identity.RemoteAddr = meta.RemoteAddr

	return &identity, nil
}
