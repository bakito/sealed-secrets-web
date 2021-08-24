package secrets

import (
	"io"
	"os/exec"
	"strings"
)

// Seal runs the kubeseal client to create the sealed secret.
func Seal(kubesealArgs string) func(secret string) ([]byte, error) {
	return func(secret string) ([]byte, error) {
		args := strings.Split(kubesealArgs, " ")
		cmd := exec.Command("kubeseal", args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, secret)
		}()

		return cmd.CombinedOutput()
	}
}
