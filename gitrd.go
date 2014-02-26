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
	"io/ioutil"
	"log"
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

	hkBytes, err := ioutil.ReadFile("hostkey")
	if err != nil {
		log.Fatalln("Failed to load private key:", err)
	}

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
