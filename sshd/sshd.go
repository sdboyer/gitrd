package sshd

import (
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
	"crypto/md5"
	"fmt"
	"github.com/sdboyer/gitrd/cfg"
	"io"
	"log"
)

type Config struct {
	// The RSA private key to use for the sshd. See ssh.ServerConfig.SetRSAPrivateKey().
	Hostkey []byte

	// The local address to which the server should bind. See net.Listen().
	BindAddr          string
	VcsRoot           string
	UserMuxing        bool
	MuxUser           string
	KeyAuthenticator  cfg.KeyAuthenticator
	PassAuthenticator cfg.PassAuthenticator
	// TODO do we really need to do KeyboardInteractive challenge mode?
	// ChallengeAuthenticator cfg.ChallengeAuthenticator
}

// Creates an ssh.ServerConfig struct that can be used to start an sshd listener.
func (c *Config) getSshServerConfig() *ssh.ServerConfig {
	config := &ssh.ServerConfig{}

	if err := config.SetRSAPrivateKey(c.Hostkey); err != nil {
		log.Fatalln("Failed to parse private key:", err)
	}

	// Set a password authenticator only if the config struct provides one.
	// Have to do it this way because the ssh package infers capability from
	// the presence or absence of the method.
	if c.PassAuthenticator != nil {
		config.PasswordCallback = func(conn *ssh.ServerConn, user, pass string) bool {
			return c.PassAuthenticator.AuthenticateUserByPassword(user, pass)
		}
	}

	// Set a pubkey authenticator only if the config struct provides one.
	if c.KeyAuthenticator != nil {
		config.PublicKeyCallback = func(conn *ssh.ServerConn, user, algo string, pubkeyBytes []byte) bool {
			if c.UserMuxing && c.MuxUser == user {
				u, err := c.KeyAuthenticator.GetUsernameFromPubkey(pubkeyBytes)
				if err != nil {
					return false
				}

				conn.User = u
				return true
			} else {
				return c.KeyAuthenticator.AuthenticateUserByPubkey(user, algo, pubkeyBytes)
			}
		}
	}

	return config
}

func Start(config *Config) *ssh.Listener {
	log.Println("Starting sshd")

	// Build an ssh.ServerConfig and start listening.
	srvcfg := config.getSshServerConfig()
	conn, err := ssh.Listen("tcp", config.BindAddr, srvcfg)
	if err != nil {
		log.Fatalf("sshd failed to listen:", err)
	}

	go func() {
		for {
			// A ServerConn multiplexes several channels, which must
			// themselves be Accepted.
			log.Println("accept")
			sConn, err := conn.Accept()
			if err != nil {
				log.Println("failed to accept incoming connection")
				continue
			}

			go func(sConn *ssh.ServerConn) {
				if err := sConn.Handshake(); err != nil {
					log.Println("failed to handshake")
				} else {
					handleServerConn(sConn)
				}
			}(sConn)
		}
	}()

	return conn
}

func getFingerprintFromKey(pubkeyBytes []byte, colons bool) (keyFingerprint string) {
	h := md5.New()
	io.WriteString(h, string(pubkeyBytes))
	if colons {
		for _, b := range h.Sum(nil) {
			keyFingerprint += fmt.Sprintf("%x:", b)
		}
		keyFingerprint = keyFingerprint[:len(keyFingerprint)-1]
	} else {
		keyFingerprint = fmt.Sprintf("%x", h.Sum(nil))
	}
	return
}

func handleServerConn(sConn *ssh.ServerConn) {
	defer sConn.Close()
	for {
		// Accept reads from the connection, demultiplexes packets
		// to their corresponding channels and returns when a new
		// channel request is seen. Some goroutine must always be
		// calling Accept; otherwise no messages will be forwarded
		// to the channels.
		ch, err := sConn.Accept()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("handleServerConn Accept:", err)
			break
		}
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if ch.ChannelType() != "session" {
			ch.Reject(ssh.UnknownChannelType, "unknown channel type")
			break
		}
		go handleChannel(ch)
	}
}

func handleChannel(ch ssh.Channel) {
	term := terminal.NewTerminal(ch, "> ")
	serverTerm := &ssh.ServerTerminal{
		Term:    term,
		Channel: ch,
	}
	ch.Accept()
	defer ch.Close()
	for {
		line, err := serverTerm.ReadLine()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("handleChannel readLine err:", err)
			continue
		}
		fmt.Println(line)
	}
}
