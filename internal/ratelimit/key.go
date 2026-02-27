package ratelimiter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net"
)

type Hasher interface {
	FromConn(conn net.Conn) (string, bool)
}

type HasherImpl struct {
	secret []byte
}

func NewHasher(secret []byte) *HasherImpl {
	return &HasherImpl{secret: secret}
}

func (h *HasherImpl) FromConn(conn net.Conn) (string, bool) {
	if conn == nil {
		return "", false
	}

	ip := extractIP(conn.RemoteAddr())
	if ip == nil {
		return "", false
	}

	return h.hash(ip)
}

func extractIP(addr net.Addr) net.IP {
	switch a := addr.(type) {

	case *net.TCPAddr:
		return normalizeIP(a.IP)

	case *net.IPAddr:
		return normalizeIP(a.IP)

	default:
		// fallback parsing
		host, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			return nil
		}
		return normalizeIP(net.ParseIP(host))
	}
}

func (h *HasherImpl) hash(ip net.IP) (string, bool) {
	mac := hmac.New(sha256.New, h.secret)
	mac.Write(ip)

	sum := mac.Sum(nil)

	// 96-bit identifier
	return hex.EncodeToString(sum[:12]), true
}

func normalizeIP(ip net.IP) net.IP {
	if ip == nil {
		return nil
	}

	// IPv4 mapped IPv6 => IPv4
	if v4 := ip.To4(); v4 != nil {
		return v4
	}

	// collapse IPv6 scanners into /64
	ip = ip.To16()
	if ip == nil {
		return nil
	}

	mask := net.CIDRMask(64, 128)
	return ip.Mask(mask)
}
