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
	UserMuxing      bool
	MuxUser         string
}

type serverConfig struct {
	ssh.ServerConfig
}

func Start(config *Config) {
	log.Println("Starting sshd")

	srvcfg := &ssh.ServerConfig{
		PasswordCallback: sshdPasswordCallback,
		PublicKeyCallback: func(conn *ssh.ServerConn, user, algo string, pubkeyBytes []byte) bool {
			if config.UserMuxing && config.MuxUser == user {
				return sshdMuxedPubkeyCallback(conn, user, algo, pubkeyBytes)
			} else {
				return sshdPubkeyCallback(conn, user, algo, pubkeyBytes)
			}
		},
	}

	pemBytes, err := ioutil.ReadFile(config.HostkeyPath)
	if err != nil {
		log.Fatalln("Failed to load private key:", err)
	}
	if err = srvcfg.SetRSAPrivateKey(pemBytes); err != nil {
		log.Fatalln("Failed to parse private key:", err)
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

		go func(sConn *ssh.ServerConn) {
			if err := sConn.Handshake(); err != nil {
				log.Println("failed to handshake")
			} else {
				handleServerConn(sConn)
			}
		}(sConn)
	}
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

func sshdPubkeyCallback(conn *ssh.ServerConn, user, algo string, pubkeyBytes []byte) bool {
	return false
}

func sshdMuxedPubkeyCallback(conn *ssh.ServerConn, user, algo string, pubkeyBytes []byte) bool {
	return false
}

func sshdPasswordCallback(conn *ssh.ServerConn, user, pass string) bool {
	return user == "testuser" && pass == "tiger"
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
