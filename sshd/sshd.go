package sshd

import (
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

//var errLog *log.Logger
//var authLog *log.Logger
//var authFailLog *log.Logger

type Config struct {
	HostkeyPath     string
	BindAddr        string
	BaseRestAddress string
	VcsRoot         string
}

type serverConfig struct {
	ssh.ServerConfig
}

func Start(config *Config) {
	log.Println("Starting sshd")

	srvcfg := &ssh.ServerConfig{
		PasswordCallback: func(conn *ssh.ServerConn, user, pass string) bool {
			return user == "testuser" && pass == "tiger"
		},
		PublicKeyCallback: func(conn *ssh.ServerConn, user, algo string, pubkeyBytes []byte) bool {
			// debug to see the pubkey as a string
			// it's base64 encoded innit: https://www.ietf.org/rfc/rfc4716.txt
			pubkeyString := base64.StdEncoding.EncodeToString(pubkeyBytes)
			fmt.Println(pubkeyString)

			// extremely dumb way to get fingerprint because i suck at Go.
			pubkeyMd5 := md5.New()
			io.WriteString(pubkeyMd5, string(pubkeyBytes))
			keyMd5 := pubkeyMd5.Sum(nil)
			keyFingerprint := ""
			for i := 0; i < len(keyMd5); i++ {
				keyFingerprint += fmt.Sprintf("%x", keyMd5[i])
				if (i + 1) < len(keyMd5) {
					keyFingerprint += ":"
				}
			}
			fmt.Println(keyFingerprint)
			// now use this to look up stuff on d.o.
			return false
		},
	}

	pemBytes, err := ioutil.ReadFile(config.HostkeyPath)
	if err != nil {
		log.Fatal("Failed to load private key:", err)
	}
	if err = srvcfg.SetRSAPrivateKey(pemBytes); err != nil {
		log.Fatal("Failed to parse private key:", err)
	}

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	conn, err := ssh.Listen("tcp", config.BindAddr, srvcfg)
	if err != nil {
		log.Fatal("failed to listen for connection")
	}
	for {
		// A ServerConn multiplexes several channels, which must
		// themselves be Accepted.
		log.Println("accept")
		sConn, err := conn.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection")
			continue
		}
		if err := sConn.Handshake(); err != nil {
			log.Println("failed to handshake")
			continue
		}
		go handleServerConn(sConn)
	}
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
