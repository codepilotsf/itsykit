package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"itsy/internal/mailer"
)

func askSecret(ctx context.Context, in io.Reader, out io.Writer) (string, error) {
	return askMasked(ctx, in, out, "Resend API key (masked)")
}

func askMasked(ctx context.Context, in io.Reader, out io.Writer, label string) (string, error) {
	timedTerminal := false
	if terminal, ok := in.(*os.File); ok {
		state := exec.Command("stty", "-g")
		state.Stdin = terminal
		mode, err := state.Output()
		if err != nil {
			return "", fmt.Errorf("Enter the API key in an interactive terminal, or set RESEND_API_KEY in your environment.")
		}
		hide := exec.Command("stty", "-echo", "-icanon", "min", "0", "time", "1")
		hide.Stdin = terminal
		if err = hide.Run(); err != nil {
			return "", fmt.Errorf("Could not hide your API key. Set RESEND_API_KEY in your environment instead.")
		}
		timedTerminal = true
		defer func() {
			restore := exec.Command("stty", strings.TrimSpace(string(mode)))
			restore.Stdin = terminal
			_ = restore.Run()
		}()
	}
	fmt.Fprint(out, label+": ")
	defer fmt.Fprintln(out)
	var key []byte
	var one [1]byte
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		n, err := in.Read(one[:])
		if err != nil {
			if err == io.EOF && timedTerminal {
				continue
			}
			if err == io.EOF && len(key) > 0 {
				return strings.TrimSpace(string(key)), nil
			}
			return "", fmt.Errorf("API key input ended before confirmation")
		}
		if n == 0 {
			continue
		}
		switch one[0] {
		case '\n', '\r':
			return strings.TrimSpace(string(key)), nil
		case 3:
			return "", context.Canceled
		case 8, 127:
			if len(key) > 0 {
				key = key[:len(key)-1]
				fmt.Fprint(out, "\b \b")
			}
		case 21:
			for len(key) > 0 {
				key = key[:len(key)-1]
				fmt.Fprint(out, "\b \b")
			}
		default:
			if one[0] >= 32 && one[0] < 127 {
				if len(key) >= 4096 {
					return "", fmt.Errorf("API key is too long")
				}
				key = append(key, one[0])
				fmt.Fprint(out, "*")
			}
		}
	}
}

func ensureMail(ctx context.Context, directory, target string, options []string, isRoot bool, in io.Reader, out, errOut io.Writer, run runner) error {
	prefix := ""
	if !isRoot {
		prefix = "sudo -- "
	}
	ssh := func(ctx context.Context, text string, tty bool, input io.Reader, output io.Writer) error {
		args := append([]string{}, options...)
		if tty {
			args = append(args, "-tt")
		} else {
			args = append(args, "-T")
		}
		args = append(args, "--", target, text)
		return run(ctx, command{name: "ssh", args: args, in: input, out: output, errOut: errOut})
	}
	var output bytes.Buffer
	if err := ssh(ctx, prefix+"/usr/local/bin/flea-server -mail-status", true, in, &output); err != nil {
		return fmt.Errorf("Could not check the server's email settings. Nothing was deployed.")
	}
	if strings.Contains(output.String(), "FLEA_MAIL=ready") {
		return nil
	}
	if !strings.Contains(output.String(), "FLEA_MAIL=missing") {
		return fmt.Errorf("The server did not report its email settings. Please update the server.")
	}
	fmt.Fprintln(out, "\nThis site needs email delivery. These settings are saved only on the server, not in Git.")
	s, err := mailer.LocalSettings(directory)
	if err != nil {
		return err
	}
	if s.APIKey == "" {
		s.APIKey, err = askSecret(ctx, in, out)
		if err != nil {
			return err
		}
	}
	if s.From == "" {
		s.From, err = ask(ctx, in, out, "Sender email (from a verified Resend domain)", "")
		if err != nil {
			return err
		}
	}
	if err = s.Validate(); err != nil {
		return err
	}
	output.Reset()
	if err = ssh(ctx, "umask 077; mktemp -d /tmp/flea-mail.XXXXXXXXXX", false, in, &output); err != nil {
		return fmt.Errorf("Could not prepare email settings on the server.")
	}
	stage := strings.TrimSpace(output.String())
	if !regexp.MustCompile(`^/tmp/flea-mail\.[A-Za-z0-9]{10}$`).MatchString(stage) {
		return fmt.Errorf("Unexpected email staging directory.")
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = ssh(cleanup, "rm -f -- "+shellQuote(stage+"/mail.json")+"; rmdir -- "+shellQuote(stage), false, nil, io.Discard)
	}()
	data, _ := json.Marshal(s)
	// No PTY while streaming secrets: a PTY could echo input into logs. sudo can
	// prompt separately when consuming the protected staged file below.
	if err = ssh(ctx, "umask 077; cat > "+shellQuote(stage+"/mail.json"), false, bytes.NewReader(data), io.Discard); err != nil {
		return fmt.Errorf("Could not transfer email settings securely.")
	}
	if err = ssh(ctx, prefix+"/usr/local/bin/flea-server -configure-mail < "+shellQuote(stage+"/mail.json"), true, in, out); err != nil {
		return fmt.Errorf("Could not save the server's email settings.")
	}
	fmt.Fprintln(out, "Email settings saved. They will take effect when this deployment restarts the server.")
	return nil
}
