package encryption

import (
	"fmt"
	"io"
	"os/exec"
)

type GPG struct {
	FilePath string
	Password string
}

func CheckForGPG() error {
	command := exec.Command(
		"gpg",
		"--version",
	)
	_, err := command.Output()
	if err != nil {
		return fmt.Errorf("could not get GPG's version: %w", err)
	}
	return nil
}

func (g GPG) EncryptFromMemoryIntoFile(data []byte) error {
	command := exec.Command(
		"gpg",
		"--batch",
		"--yes",
		"--passphrase-fd=0",
		"--symmetric",
		"--armor",
		"--output",
		g.FilePath,
	)
	pipe, err := command.StdinPipe()
	if err != nil {
		return fmt.Errorf("could not create a pipe to the GPG command's input: %w", err)
	}
	err = command.Start()
	if err != nil {
		return fmt.Errorf("could not start the GPG command: %w", err)
	}

	_, err = pipe.Write(([]byte)(fmt.Sprintf("%s\n%s", g.Password, (string)(data))))
	if err != nil {
		return fmt.Errorf("could not write into the GPG command's input: %w", err)
	}
	err = pipe.Close()
	if err != nil {
		return fmt.Errorf("could not close the pipe to the GPG command's input: %w", err)
	}

	err = command.Wait()
	if err != nil {
		return fmt.Errorf("could not wait for the GPG command: %w", err)
	}
	return nil
}

// TODO: Return better errors when the password is invalid.
func (g GPG) DecryptFromFileIntoMemory() ([]byte, error) {
	command := exec.Command(
		"gpg",
		"--batch",
		"--yes",
		"--passphrase-fd=0",
		"--decrypt",
		g.FilePath,
	)
	inPipe, err := command.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not create a pipe to the GPG command's input: %w", err)
	}
	outPipe, err := command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not create a pipe to the GPG command's output: %w", err)
	}
	err = command.Start()
	if err != nil {
		return nil, fmt.Errorf("could not start the GPG command: %w", err)
	}

	_, err = inPipe.Write(([]byte)(fmt.Sprintf("%s\n", g.Password)))
	if err != nil {
		return nil, fmt.Errorf("could not write into the GPG command's input: %w", err)
	}
	err = inPipe.Close()
	if err != nil {
		return nil, fmt.Errorf("could not close the pipe to the GPG command's input: %w", err)
	}

	data, err := io.ReadAll(outPipe)
	if err != nil {
		return nil, fmt.Errorf("could not read from the GPG command's output: %w", err)
	}
	err = outPipe.Close()
	if err != nil {
		return nil, fmt.Errorf("could not close the pipe to the GPG command's output: %w", err)
	}

	err = command.Wait()
	if err != nil {
		return nil, fmt.Errorf("could not wait for the GPG command: %w", err)
	}
	return data, nil
}
