package sshd

type Identity struct {
	User              string
	UserID            string
	PublicKey         []byte
	Metadata          map[string]string
	RemoteAddr        string
	RequestedHostname string
}

type Pty struct {
	Term string
	Cols int
	Rows int
}

type ResizeEvent struct {
	Width  int
	Height int
}
