package sshd

// Tests ensuring that configuration-driven options translate to correct
// behaviors.

import (
	"code.google.com/p/go.crypto/ssh"
	"testing"
)

func TestConfigFromConfig(t *testing.T) {
	our_config := &Config{
		Hostkey:  []byte(testServerPrivateKey),
		BindAddr: "127.0.0.0:0",
		VcsRoot:  "repos",
	}

	a := auther(true)
	var cfg *ssh.ServerConfig

	cfg = our_config.getSshServerConfig()

	if cfg.PublicKeyCallback != nil {
		t.Error("Pubkey callback was bound, but no implementation was provided. sshd will erroneously offer pubkey auth.")
	}

	if cfg.PasswordCallback != nil {
		t.Error("Password callback was bound, but no implementation was provided. sshd will erroneously offer pasword auth.")
	}

	our_config.KeyAuthenticator = a
	cfg = our_config.getSshServerConfig()

	if cfg.PublicKeyCallback == nil {
		t.Error("Pubkey callback was not bound, but an implementation was provided. sshd will erroneously not offer pubkey auth.")
	}

	if cfg.PasswordCallback != nil {
		t.Error("Password callback was bound, but no implementation was provided. sshd will erroneously offer pasword auth.")
	}

	our_config.PassAuthenticator = a
	cfg = our_config.getSshServerConfig()

	if cfg.PublicKeyCallback == nil {
		t.Error("Pubkey callback was not bound, but an implementation was provided. sshd will erroneously not offer pubkey auth.")
	}

	if cfg.PasswordCallback == nil {
		t.Error("Password callback was not bound, but an implementation were provided. sshd will erroneously not offer pasword auth.")
	}
}
