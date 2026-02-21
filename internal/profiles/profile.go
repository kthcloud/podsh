package profiles

import (
	"context"
	"fmt"
	"strings"

	"github.com/kthcloud/podsh/internal/server"
	"github.com/spf13/viper"
)

var profiles = map[ProfileKey]Profile{
	ProfileKeyDev:  DevProfileImpl{},
	ProfileKeyProd: nil,
}

type Profile interface {
	Mode() Mode
	Config(ctx context.Context, v *viper.Viper) (*server.Config, error)
}

type Mode = uint8

const (
	ModeDev Mode = iota
	ModeProd
)

func Get(key ProfileKey) Profile {
	return profiles[key]
}

type ProfileKey string

const (
	ProfileKeyDev  ProfileKey = "dev"
	ProfileKeyProd ProfileKey = "prod"
)

func allowedProfileKeyString() string {
	return strings.Join([]string{string(ProfileKeyDev), string(ProfileKeyProd)}, ", ")
}

func (p ProfileKey) IsValid() bool {
	switch p {
	case ProfileKeyDev, ProfileKeyProd:
		return true
	}
	return false
}

type ProfileFlag struct {
	Value *ProfileKey
}

func (e *ProfileFlag) String() string {
	if e.Value == nil {
		return ""
	}
	return string(*e.Value)
}

func (e *ProfileFlag) Set(s string) error {
	p := ProfileKey(strings.ToLower(s))

	if !p.IsValid() {
		return fmt.Errorf("invalid profile '%s' (allowed: %s)", s, allowedProfileKeyString())
	}

	*e.Value = p
	return nil
}

func (e *ProfileFlag) Type() string {
	return "profile"
}

func NewProfileFlag(defaultVal ProfileKey) (*ProfileFlag, ProfileKey) {
	v := defaultVal
	return &ProfileFlag{Value: &v}, v
}
