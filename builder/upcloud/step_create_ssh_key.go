package upcloud

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"golang.org/x/crypto/ssh"
)

const (
	sshKeyBitSize = 3072
)

// StepCreateSSHKey represents the step that creates ssh private and public keys.
type StepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

// Run runs the actual step.
func (s *StepCreateSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui, ok := state.Get("ui").(packer.Ui)
	if !ok {
		return stepHaltWithError(state, errors.New("UI is not of expected type"))
	}
	config, ok := state.Get("config").(*Config)
	if !ok {
		return stepHaltWithError(state, errors.New("config is not of expected type"))
	}

	if config.SSHPrivateKeyPath != "" && config.SSHPublicKeyPath != "" {
		var err error
		ui.Say("Using provided SSH keys...")

		if config.Comm.SSHPrivateKey, err = os.ReadFile(config.SSHPrivateKeyPath); err != nil {
			return stepHaltWithError(state, fmt.Errorf("failed to read private key: %w", err))
		}

		if config.Comm.SSHPublicKey, err = os.ReadFile(config.SSHPublicKeyPath); err != nil {
			return stepHaltWithError(state, fmt.Errorf("failed to read public key: %w", err))
		}

		state.Put("ssh_key_public", strings.Trim(string(config.Comm.SSHPublicKey), "\n"))
		return multistep.ActionContinue
	}

	ui.Say("Creating temporary ssh key...")

	priv, err := rsa.GenerateKey(rand.Reader, sshKeyBitSize)
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("error generating SSH key: %w", err))
	}

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(priv)
	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}

	// Marshal the public key into SSH compatible format
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("error creating public ssh key: %w", err))
	}

	// Remember some state for the future
	pubSSHFormat := string(ssh.MarshalAuthorizedKey(pub))
	state.Put("ssh_key_public", pubSSHFormat)

	// Set the private key in the config for later
	config.Comm.SSHPrivateKey = pem.EncodeToMemory(&privBlk)
	config.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)

	// If we're in debug mode, output the private key to the working directory.
	if s.Debug {
		ui.Say(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		err := os.WriteFile(s.DebugKeyPath, config.Comm.SSHPrivateKey, 0o600)
		if err != nil {
			return stepHaltWithError(state, fmt.Errorf("error saving debug key: %w", err))
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {}
