package services

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHServiceImpl implements SSHService
type SSHServiceImpl struct{}

// NewSSHService creates an SSHService
func NewSSHService() SSHService {
	return &SSHServiceImpl{}
}

func (s *SSHServiceImpl) AddKey(ctx context.Context, hostname string, port int, sshUser, privateKey, publicKey string) error {
	client, session, err := s.connect(ctx, hostname, port, sshUser, privateKey)
	if err != nil {
		slog.Error("SSH connect failed for AddKey", "hostname", hostname, "port", port, "error", err)
		return err
	}
	defer client.Close()
	slog.Debug("SSH AddKey", "hostname", hostname, "port", port)

	publicKey = strings.TrimSpace(publicKey)
	if publicKey == "" {
		return fmt.Errorf("empty public key")
	}

	cmd := fmt.Sprintf("mkdir -p ~/.ssh && chmod 700 ~/.ssh && grep -qF '%s' ~/.ssh/authorized_keys || echo '%s' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys",
		strings.ReplaceAll(publicKey, "'", "'\"'\"'"),
		strings.ReplaceAll(publicKey, "'", "'\"'\"'"),
	)
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if err := session.Run(cmd); err != nil {
		slog.Error("SSH AddKey run failed", "hostname", hostname, "error", err)
		return fmt.Errorf("ssh add key: %w: %s", err, stderr.String())
	}
	return nil
}

func (s *SSHServiceImpl) RemoveKey(ctx context.Context, hostname string, port int, sshUser, privateKey, publicKey string) error {
	client, session, err := s.connect(ctx, hostname, port, sshUser, privateKey)
	if err != nil {
		slog.Error("SSH connect failed for RemoveKey", "hostname", hostname, "port", port, "error", err)
		return err
	}
	defer client.Close()
	slog.Debug("SSH RemoveKey", "hostname", hostname, "port", port)

	publicKey = strings.TrimSpace(publicKey)
	if publicKey == "" {
		return fmt.Errorf("empty public key")
	}

	// Use grep -v to remove the line containing the key
	escaped := strings.ReplaceAll(publicKey, "'", "'\"'\"'")
	cmd := fmt.Sprintf("grep -vF '%s' ~/.ssh/authorized_keys > ~/.ssh/authorized_keys.tmp 2>/dev/null && mv ~/.ssh/authorized_keys.tmp ~/.ssh/authorized_keys || true", escaped)
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	_ = session.Run(cmd)
	return nil
}

func (s *SSHServiceImpl) TestConnection(ctx context.Context, hostname string, port int, sshUser, privateKey string) error {
	client, _, err := s.connect(ctx, hostname, port, sshUser, privateKey)
	if err != nil {
		slog.Error("SSH TestConnection failed", "hostname", hostname, "port", port, "error", err)
		return err
	}
	client.Close()
	slog.Debug("SSH TestConnection success", "hostname", hostname, "port", port)
	return nil
}

func (s *SSHServiceImpl) connect(ctx context.Context, hostname string, port int, sshUser, privateKey string) (*ssh.Client, *ssh.Session, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, nil, fmt.Errorf("parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil // skip host key verification for simplicity
		},
		Timeout: 10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, nil, fmt.Errorf("ssh dial: %w", err)
	}

	session, err := conn.NewSession()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("new session: %w", err)
	}

	return conn, session, nil
}
