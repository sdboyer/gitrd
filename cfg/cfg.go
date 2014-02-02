// Defines config and general datastructures for use in gitrd.
package cfg


type Repository struct {
	Path     string
	Name     string
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
	Name     string
	Uid      int
	Keys     [][]byte
	Password []byte
	Auth     Authorizor
}

type UserCache map[string]*User

type Authorizor interface {
	CanRead(User, Repository) (bool, string)
	CanWrite(User, Repository) (bool, string)
}

