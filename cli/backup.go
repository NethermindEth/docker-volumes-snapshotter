package cli

import (
	"github.com/NethermindEth/docker-volumes-snapshotter/internal/backup"
	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/config"
	"github.com/spf13/cobra"
)

func BackupCmd() *cobra.Command {
	return &cobra.Command{
		Use: "backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadConfig()
			if err != nil {
				return err
			}
			err = backup.Backup(conf)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
