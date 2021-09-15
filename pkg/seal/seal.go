package seal

import (
	"errors"
	"io"
	"os/exec"
	"strings"
)

type Sealer interface {
	Secret(secret string) ([]byte, error)
	Raw(data Raw) ([]byte, error)
}

// New create a new sealer
func New(args []string) Sealer {
	return &sealer{
		args: args,
	}
}

type sealer struct {
	args []string
}

func (s *sealer) Secret(secret string) ([]byte, error) {
	return s.kubeseal(secret)
}

func (s *sealer) Raw(data Raw) ([]byte, error) {
	args := []string{"--raw", "--name", data.Name}
	if strings.TrimSpace(data.Namespace) != "" {
		args = append(args, "--namespace", data.Namespace)
	}
	if strings.TrimSpace(data.Scope) != "" {
		args = append(args, "--scope", data.Scope)
	}
	return s.kubeseal(data.Value, args...)
}

func (s *sealer) kubeseal(value string, additionalArgs ...string) ([]byte, error) {
	args := s.args
	args = append(args, additionalArgs...)
	cmd := exec.Command("kubeseal", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()
		_, _ = io.WriteString(stdin, value)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.New(strings.Replace(strings.TrimSpace(string(out)), "error: ", "", 1))
	}
	return out, err
}

type Raw struct {
	Value     string `json:"value"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Scope     string `json:"scope"`
}
