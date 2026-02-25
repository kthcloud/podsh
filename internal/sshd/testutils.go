package sshd

import (
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// NewTestServerConfig returns an ssh.ServerConfig with a test host key
func NewTestServerConfig() *ssh.ServerConfig {
	privateKey, err := ssh.ParsePrivateKey([]byte(testServerPrivateKey))
	if err != nil {
		panic("failed to parse server key: " + err.Error())
	}

	serverConfig := &ssh.ServerConfig{
		NoClientAuth: true, // allow auth-less for testing
	}
	serverConfig.AddHostKey(privateKey)
	return serverConfig
}

// NewTestClientConfig returns a simple ssh.ClientConfig for testing
func NewTestClientConfig(user string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
}

// NewPipePair returns a connected in-memory client/server net.Conn pair
func NewPipePair() (serverConn net.Conn, clientConn net.Conn) {
	return net.Pipe()
}

// Test server private key (PEM)
const testServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAuZtCgZ0rS5YI4p/5FAZ1Tgq7ZsKIpQmGNH7XbWxhNSjvxh0n
x4n7S4ILxChT+zv7QbDsqz2rxh/8x3kakD7+MHL1GVjyzCLt03G9WyVQ2JuVvPjk
2Fey2XxGZ0F4+IdkQb2Epo0HUl3FGb0HPU0m6VlnhC4gpiYHGmFXEcsSZyE7tNml
kRfXQf0zD7x1I4MSkg3ZZRV7W4eBDYvF4f/tpcKexai3B6D8gHhNxD0xg1F9g+Lc
Wn3m+Y9/0S18NsEzQl0qNHDXoZ2kFONu7xRmvY8fX+3BvC2Fg2HkM8zO+pJUrGHV
uQ8Sxg7qv6x7eGZ+Z7KH9BcdChU3Xip8nZ2MvwIDAQABAoIBAGq4r9YtR15b1KXy
dCdd0jH0SxEYxq3V7evkqehcQ3IDp2nF84Q2D+P0v1C6qq3HtmZzAOKqMX+1iwWr
Lw+z3mmHVVScE+9sGfV8u6y7GpV8c6ZPfxygQ2yYc3rkuZrpttL5P4jHG6bFG+N/
7gIRP6a+v+6ZhU7NFRN+YzF5k5p6V2DzXq7s6uWjYefw/ZK0JxDLp3sw2MEC+K3k
qJOKlF+QKQzF0D4Y+9J4iR0r+63t0Z8S3ytE8w0Yd9Q6U8JtI9Q5jJk7g+Z1x6Ov
2Y7L+MXjD+u/AZCtwgq7YjYrP7SnlGn/gdZBG5VYjYwzdp2v0wFXeC9Rk/8QfSgV
Mj6XxSECgYEA8WwV+SyV7z+exVr4+Kj5l8aZk4u+R3qu5JXKd+Lr2k0+lWcLMf37
G8h94zqjh/5U3+8Q6eH0KrZC8yGeT3V7qS7A/8k1uZqHT+7PpMNQv7twj5FqUpJ1
v3nm1M9rYqUfG+ryxH5Vz4xLqj5Zg7M1RU2v9n0gJc8jHv9hF3oi0p/sCgYEAy9Q
nM7TQd0K1tF/1zNkmRZ37U4nD3kE/xjtt6H2bg17+5JXhFJxGqDpJz7Jv5H2D+O8
1TVn5Oq5/SXx1k7Xk3/j6n0K9j6vU1uF5bWhmDD6B6R9bVrkYwJKUsc7I6HvGZyc
9O0GqVAvU1sVxZfEyx7pR7Lrnq+7FynCkYLUv8MCgYEAhz4L29TfxHTv5Oft0wGB
jN9WgRjR8ukAly1XcVhQ/ajkIVbdexhF5TrLwPqAhXk9PKNvd0M6sZ7Fm+gWlB1V
F5tjILj6u2+0o6kL5bM4AwZ17kR0L4cI7H4E1N8KpHjKq7nD8Zx7Ff3zZp7S6Z1X
9F42d1sCgYEA0c9vXm1XyWwQJYxk7TX75sjDzC7Jrf4U6NdrfhYc+uv0cOQkDgPR
BKz0GeE8uWmU9L3c7qD0iXkDkV/wH9V0S6gUZkvlWAlkOdG+GHsnzY0pS6wPzHbd
HfL1TVkrn/N2w0C+2lbL7S6xJ8NckVYk1JzVtPku9DE=
-----END RSA PRIVATE KEY-----`
