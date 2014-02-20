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

type Authorizor interface {
	CanRead(User, Repository) (bool, string)
	CanWrite(User, Repository) (bool, string)
}

type KeyAuthenticator interface {
	// Given a pubkey, determine which user it belongs to (if any). This is
	// used only if ssh user multiplexing is enabled and the multiplexing user
	// has been provided.
	//
	// If no user is found for the pubkey, an error is returned. May be called
	// concurrently from several goroutines.
	GetUsernameFromPubkey(pubkeyBytes []byte) (username string, err error)

	// Given a pubkey and a username, indicates if the pubkey is valid for
	// that user. May be called concurrently from several goroutines.
	AuthenticateUserByPubkey(user, algo string, pubkeyBytes []byte) (valid bool)
}

type PassAuthenticator interface {
	// Given a user and a plaintext password, indicates if the password is
	// valid for that user. May be called concurrently from several goroutines.
	AuthenticateUserByPassword(user, pass string) (valid bool)
}
