package cmd

import (
	"github.com/gpitfield/filmstrip/build"
	"github.com/gpitfield/filmstrip/deploy"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "filmstrip",
	Short: "Generate and deploy the filmstrip site",
	Run: func(cmd *cobra.Command, args []string) {
		build.Build()
		deploy.Deploy()
	},
}

var dpl = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the 'site' folder.",
	Run: func(cmd *cobra.Command, args []string) {
		deploy.Deploy()
	},
}

var bld = &cobra.Command{
	Use:   "build",
	Short: "Generate the 'site' folder.",
	Run: func(cmd *cobra.Command, args []string) {
		build.Build()
	},
}

func init() {
	RootCmd.AddCommand(dpl)
	RootCmd.AddCommand(bld)
}
