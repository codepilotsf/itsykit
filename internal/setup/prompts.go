package setup

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// Prompt reads a cancellable terminal answer using the same rules as setup.
func Prompt(ctx context.Context, in io.Reader, out io.Writer, label string) (string, error) {
	return ask(ctx, in, out, label, "")
}

// Read exactly one answer. Do not buffer input needed by a later SSH password
// prompt or another command in the same interactive flow.
func ask(ctx context.Context, in io.Reader, out io.Writer, label, def string) (string, error) {
	fmt.Fprint(out, label)
	if def != "" {
		fmt.Fprintf(out, " [%s]", def)
	}
	fmt.Fprint(out, ": ")
	type answer struct {
		line string
		err  error
	}
	done := make(chan answer, 1)
	go func() {
		var b strings.Builder
		var one [1]byte
		for {
			n, err := in.Read(one[:])
			if n > 0 {
				if one[0] == '\n' {
					done <- answer{line: b.String()}
					return
				}
				b.WriteByte(one[0])
			}
			if err != nil {
				done <- answer{line: b.String(), err: err}
				return
			}
		}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case a := <-done:
		if a.err != nil && !(a.err == io.EOF && a.line != "") {
			return "", fmt.Errorf("Please run this command in a terminal so you can answer the setup questions.")
		}
		value := strings.TrimSpace(a.line)
		if value == "" {
			value = def
		}
		return value, nil
	}
}

func sshUsername(ctx context.Context, target string, in io.Reader, out io.Writer) (string, error) {
	normalized, err := targetName(target)
	if err != nil {
		return "", err
	}
	if strings.Contains(normalized, "@") {
		return normalized, nil
	}
	user, err := ask(ctx, in, out, "SSH username", "root")
	if err != nil {
		return "", err
	}
	return targetName(user + "@" + normalized)
}

func confirmSetup(ctx context.Context, in io.Reader, out io.Writer) (bool, error) {
	for {
		value, err := ask(ctx, in, out, "Install Flea CMS on this server now? (y/n)", "y")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(value) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
		fmt.Fprintln(out, "Please enter y or n.")
	}
}
