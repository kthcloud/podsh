// Package requests contain structs for different request payloads.
package requests

// PTYReq represents the payload sent with the pty-req request,
// Format has been determined from:
// https://datatracker.ietf.org/doc/html/rfc4254#section-6.2
type PTYReq struct {
	TERM     string // TERM environment variable value
	Cols     uint32 // terminal width in characters
	Rows     uint32 // terminal height in rows
	WidthPx  uint32 // terminal width in pixels
	HeightPx uint32 // terminal height in pixels
	Mode     string // encoded terminal modes
}

// DirectTCPIP represents the payload sent with the direct-tcpip request,
// https://datatracker.ietf.org/doc/html/rfc4254#section-7.2
type DirectTCPIP struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

// ExecRequest represents the payload sent with the exec request,
type ExecRequest struct {
	Command string
}

// SubsystemRequest represents the payload sent with the subsystem request,
type SubsystemRequest struct {
	Subsystem string
}

// WindowChangeRequest represents the payload sent with the window-change requets
type WindowChangeRequest struct {
	Cols         uint32
	Rows         uint32
	WidthPixels  uint32
	HeightPixels uint32
}

// EnvRequest represents the payload sent with the enc requet
type EnvRequest struct {
	Name  string // variable name
	Value string // variable value
}
