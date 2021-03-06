/*
gitrd is a daemon to serve all your remote git needs.

It handles SSH, (Smart) HTTP, and at some point, maybe even the git protocol,
all through a single daemon. It also exposes a RESTful interface for accessing
data about the git repository that is not available with conventional git.

By combining all these behaviors into a single daemon, it simplifies the task
of running a large git hosting service. gitrd handles routing, ident, auth
and other configurable components to be managed through a single stack,
eliminating the need to teach multiple different daemons about how your git
infrastructure works.
*/
package main

import (
	//"github.com/jessevdk/go-flags"
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"github.com/sdboyer/do_git_rest/keys"
	"github.com/sdboyer/gitrd/sshd"
	//"log"
)

// TODO temporary approach to blocking the main process once sshd is spawned
var blockerchan = make(chan int)

type baseOpts struct {
	Verbose bool `short:"v" long:"verbose" description:"enables verbose output"`
	Quiet   bool `short:"q" long:"quiet" description:"turns off all output"`
	// TODO figure out how to handle/reconcile a config dir with go-flags options.
}

var defaultOpts = &baseOpts{
	Verbose: false,
	Quiet:   false,
}

func main() {
	/*
	   p := flags.NewParser(defaultOpts, flags.HelpFlag|flags.PrintErrors)
	   	p.Usage = `[OPTIONS] ...

	   	gitrd is an all-in-one git daemon: ssh, http, etc.`
	*/

	hkBytes := []byte(`-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`)

	a := auther(true)
	ssh_config := &sshd.Config{
		Hostkey:           hkBytes,
		BindAddr:          "0.0.0.0:2022",
		VcsRoot:           "repos",
		UserMuxing:        true,
		MuxUser:           "git",
		KeyAuthenticator:  a,
		PassAuthenticator: a,
	}

	sshd.Start(ssh_config)

	for {
		// Just sit and block here, for now
		<-blockerchan
	}
}

type auther bool

func (a auther) GetUsernameFromPubkey(pubkeyBytes []byte) (username string, err error) {
	for username, key := range keys.Keydata {
		pubkey := key.PublicKey()

		if bytes.Equal(pubkeyBytes, ssh.MarshalPublicKey(pubkey)) {
			return username, nil
		}
	}

	return "", errors.New("No user found with the given pubkey.")
}

func (a auther) AuthenticateUserByPubkey(user, algo string, pubkeyBytes []byte) (valid bool) {
	key, exists := keys.Keydata[user]
	if !exists {
		return false
	}

	pubkey := key.PublicKey()
	return algo == pubkey.PublicKeyAlgo() && bytes.Equal(pubkeyBytes, ssh.MarshalPublicKey(pubkey))
}

func (a auther) AuthenticateUserByPassword(user, pass string) (valid bool) {
	return bool(a)
}
