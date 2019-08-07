package secrets

import (
	"io"
	"os"
	"os/exec"
)

// Seal runs the kubeseal client to create the sealed secret.
func Seal(secret string) ([]byte, error) {
	cmd := exec.Command("kubeseal", os.Args[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, secret)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return out, err
}
