package cli

import (
	"github.com/NethermindEth/docker-volumes-snapshotter/internal/backup"
	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/config"
	"github.com/spf13/cobra"
)

func BackupCmd() *cobra.Command {
	var configFlag string
	cmd := cobra.Command{
		Use: "backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadConfig(configFlag)
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
	cmd.Flags().StringVar(&configFlag, "config", "/snapshotter.yml", "config file with all the volumes to save")
	return &cmd
}
