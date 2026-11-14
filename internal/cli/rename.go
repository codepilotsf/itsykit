package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"flea/internal/buildinfo"
)

// RunExecutable keeps the retired command separate from normal CLI dispatch.
func RunExecutable(ctx context.Context, executable string, args []string, out, errOut io.Writer) int {
	name := filepath.Base(executable)
	// Released Flea updaters verify a randomly named candidate before replacing
	// their executable. Keep this exact, read-only response for that probe only.
	if strings.HasPrefix(name, ".flea-update-") && len(args) == 1 && args[0] == "version" {
		fmt.Fprintln(out, "FleaCMS "+buildinfo.Version)
		return 0
	}
	if name == "flea" {
		if len(args) == 1 && args[0] == "update" {
			fmt.Fprint(errOut, "\nFLEA HAS CHANGED TO ITSYWEB\n\nThe CLI command is now itsy. This update installs itsy on this computer.\nFrom now on, please use itsy. For future updates, run itsy update.\n\n")
			return Run(ctx, args, out, errOut)
		}
		fmt.Fprint(errOut, "Flea is now ItsyWeb. From now on, please use itsy.\nIf itsy is not installed yet, run flea update.\n")
		return 2
	}
	return Run(ctx, args, out, errOut)
}
