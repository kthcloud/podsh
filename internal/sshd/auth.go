package sshd

import "context"

type PublicKeyAuthenticator interface {
	Authenticate(ctx context.Context, meta ConnMetadata, pubKey []byte) (*Identity, error)
}

type ConnMetadata struct {
	User       string
	RemoteAddr string
}
