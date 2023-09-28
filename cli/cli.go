package cli

import (
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "snapshotter",
	}

	cmd.AddCommand(BackupCmd())
	cmd.AddCommand(RestoreCmd())

	return &cmd
}
