package cli

import (
	"github.com/NethermindEth/docker-volumes-snapshotter/internal/backup"
	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/config"
	"github.com/spf13/cobra"
)

func RestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use: "restore",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadConfig()
			if err != nil {
				return err
			}
			err = backup.Restore(conf)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
