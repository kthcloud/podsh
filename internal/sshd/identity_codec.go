package sshd

import (
	"encoding/base64"
	"encoding/json"
)

const identityVersion = 1

type wireIdentity struct {
	Version int               `json:"v"`
	User    string            `json:"u"`
	UserID  string            `json:"i"`
	Key     []byte            `json:"k"`
	Meta    map[string]string `json:"m,omitempty"`
	Addr    string            `json:"a,omitempty"`
}

func encodeIdentity(id *Identity) string {
	if id == nil {
		return ""
	}

	w := wireIdentity{
		Version: identityVersion,
		User:    id.User,
		UserID:  id.UserID,
		Key:     id.PublicKey,
		Meta:    id.Metadata,
		Addr:    id.RemoteAddr,
	}

	b, err := json.Marshal(w)
	if err != nil {
		return ""
	}

	return base64.RawStdEncoding.EncodeToString(b)
}

func decodeIdentity(raw string) Identity {
	if raw == "" {
		return Identity{}
	}

	data, err := base64.RawStdEncoding.DecodeString(raw)
	if err != nil {
		return Identity{}
	}

	var w wireIdentity
	if err := json.Unmarshal(data, &w); err != nil {
		return Identity{}
	}

	// future-proofing
	if w.Version != identityVersion {
		return Identity{}
	}

	return Identity{
		User:       w.User,
		UserID:     w.UserID,
		PublicKey:  w.Key,
		Metadata:   w.Meta,
		RemoteAddr: w.Addr,
	}
}
