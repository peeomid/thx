package cli

import "github.com/spf13/cobra"

func addPagingFlags(cmd *cobra.Command, limit *int, offset *int) {
	cmd.Flags().IntVar(limit, "limit", 0, "limit number of results")
	cmd.Flags().IntVar(offset, "offset", 0, "offset results")
}
