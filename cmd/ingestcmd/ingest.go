package ingestcmd

import (
	"github.com/spf13/cobra"
)

func init() {
	IngestCmd.AddCommand(ingestMonthlyRepositoryDataCmd)
}

var IngestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest data from the source systems",
}
