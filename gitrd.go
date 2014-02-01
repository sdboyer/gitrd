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
package gitrd

import (
	"github.com/jessevdk/go-flags"
)

type baseOpts struct {
	Verbose bool `short:"v" long:"verbose" description:"enables verbose output"`
	Quiet   bool `short:"q" long:"quiet" description:"turns off all output"`
	// TODO figure out how to handle/reconcile a config dir with go-flags options.
}

var defaultOpts = &baseOpts {
	Verbose: false,
	Quiet:   false,
}

type Repository struct {
	Path string
	Name string
	Disabled bool
}

type PusherChan struct {
	channel chan []byte
}

type PullerChan struct {
	channel chan []byte
}

// This would be the thing we call once we determine that a push is
// desired by the client. Should be able to be called from httpd or
// sshd context, if possible.
func (r *Repository) ReceivePack(filler PusherChan) {
}

// This would be the thing we call once we determine that a pull is
// desired by the client. Should be able to be called from httpd or
// sshd context, if possible.
func (r *Repository) UploadPack(filler PullerChan) {
}

type User struct {
	Name string
	Uid int
	Keys [][]byte
	Password []byte
	Auth Authorizor
}

type UserCache map[string]*User

type Authorizor interface {
	CanRead(User, Repository) (bool, string)
	CanWrite(User, Repository) (bool, string)
}

func main() {
	p := flags.NewParser(defaultOpts, flags.HelpFlag|flags.PrintErrors)
	p.Usage = `[OPTIONS] ...

	gitrd is an all-in-one git daemon: ssh, http, etc.`
}
