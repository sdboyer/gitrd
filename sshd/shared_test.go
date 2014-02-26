package sshd

// Shared functions, fixtures, etc., for other sshd tests.

import (
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"testing"
)

const testServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA4W0/KcZiLTWC7oCxCxlof1MLihRV3e0bXvSfcC8dVm2roi/x
6M4b2yzfaSXgoE62ReCPSxAjXw9bqTlJtKvuCq0q54rSxZ26IWtD+G4wTASU2sNe
6tqhHWdQ6FrpQV7pyzwKWx9iheLdaiw0OecW7DQi1LMxgW4cjinApJIDev3JSmPZ
PzOJmkqi9g5NuiJetrAEKhD3APAzfvpMmVKgJxs31+tRUjoYt1df5MLKJJueHWwv
xWV3bHyXNY668m+VlK08SO5zwUpnzsX5wQMQndxk/Hbb03zXfLg5Jpq2/hq9ZlLE
HPCMsvaCqLK05GJLgj9k18QacRsnE3eKB8997QIDAQABAoIBAQDSw2ucyUidcDyc
dWISOI1FDgXp8Z1ewwMmQpyXLNXHKv6fwyfwPFQ7FbdD/hAIkc9FgfE3gz0u8ZMH
ovJQo7cJ8GH+3gK2lQOjn0CLk6pASMBL0QJ7njGo5iH1PJp9bho01EvyamOZPkU5
sV6bDH6YFR7Ds06D7slv+YWN2J68a635zkVVdZ82xPu1J2lnx70kjBznoiT3A8CQ
I1Qw/2Fp3vi4JR2vLIk62MMPxO52qUwBWHfLMBz6gbUOO9mrocMMEXk0+Ytl4izK
CT+w9XHYP8E2xsmbSMjUAwhCvwNO/+hSVzJi7BnF9xbd5cu9D6l72mvWtF4JVvmW
k4vDn10BAoGBAPxKr3hyqa4J4HRG2xhbTzGfL1y/WnvoKXScEKsWUa+1Ao3fRSgh
0wJ1fnVX+2R1teVuMjNkJ5wsHK0tVbrxw05THY0GWUly/Ti8v+Ru8BaCUgEesnw2
QX7RQnkBvisM0286PMIy/D4OKFL/axA7LsAIwZ4xy7WtasjOH8++tWYNAoGBAOS9
eYGHu2FdnCpY2o7sj7kvj/S36sLiy09LLZi0B5J2J9xi3JmxG5s4Eim1w8DH8OYa
wbkMo3BEpmgOYrJHMv9yNQrwcwV9+2shvbbBgxwkeYe1/jHdgx+HWssTDt/CtLqU
q6ilrzv+PDAxIKGIQan+DMN6CttpQ/JKEUCkyV9hAoGALpGWnBQGMALQtIXTsUZB
cvZgJq2HhTGQXV7lUL846sbtpsRcnpDHwz9uzTglRiDYJ3ZUu9mz2gbmcCzbEzvH
AjEjVkGiv4UDKrLkdMTpei4p9tz0syrMohz8ORvSP14JtRE539rLZqT0WoWc/I0A
DyBOpOWqJWnSOSibBJy+HQ0CgYEA322QVQzO5EE6vEaEXd0GWj3yIHjRkEFVlAN7
60/WoaJWNzg+AMX1kD1JyIIqTpE+ZpU2KtoEfzIfVT7P+xH+53OYCjJqN7AiODgC
BpSoy4F5UC1duTmEzfQ5pGjeO4UFYca8kgQc0b347p3eIMpmUXS85Oe92SnOW8kr
ZvhPVqECgYBZ0C1L1cqygt5et/5vWIMTvTCEDv4agJ5pikx2Hs7Ahg5XlNJ3otkB
hsG5iWLE46k0gD7EK6Wpt6/oePKA/RH7HXodfX7+Eig+Rq2tcGFgR+Vzb+l2cXH5
sr0cXBFl9v3W/Fta1W/sf9AzN17vMdqAt6Jxa1POab8Sw5leNN/vNw==
-----END RSA PRIVATE KEY-----`

type auther bool

func (a auther) GetUsernameFromPubkey(pubkeyBytes []byte) (username string, err error) {
	if bool(a) {
		return "keyuser", nil
	} else {
		return "", errors.New("No user known for provided pubkey.")
	}
}

func (a auther) AuthenticateUserByPubkey(user, algo string, pubkeyBytes []byte) (valid bool) {
	return bool(a)
}

func (a auther) AuthenticateUserByPassword(user, pass string) (valid bool) {
	return bool(a)
}

// Creates a mock auth server that exits after one handshake. Modelled after
// a similar system in go.crypto/ssh/client_auth_test.go.
func createMockAuthServer(t *testing.T, c *Config) string {
	l, err := ssh.Listen("tcp", c.BindAddr, c.getSshServerConfig())
	if err != nil {
		t.Fatalf("unable to newMockAuthServer: %s", err)
	}
	go func() {
		defer l.Close()
		c, err := l.Accept()
		if err != nil {
			t.Errorf("Unable to accept incoming connection: %v", err)
			return
		}
		if err := c.Handshake(); err != nil {
			// not Errorf because this is expected to
			// fail for some tests.
			t.Logf("Handshaking error: %v", err)
			return
		}
		defer c.Close()
	}()
	return l.Addr().String()
}
