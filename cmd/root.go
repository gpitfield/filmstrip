package cmd

import (
	"github.com/gpitfield/filmstrip/build"
	"github.com/gpitfield/filmstrip/deploy"
	"github.com/spf13/cobra"
)

var force bool

var RootCmd = &cobra.Command{
	Use:   "filmstrip",
	Short: "Generate and deploy the filmstrip site",
	Run: func(cmd *cobra.Command, args []string) {
		build.Build(force)
		deploy.Deploy(force)
	},
}

var dpl = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the 'site' folder.",
	Run: func(cmd *cobra.Command, args []string) {
		deploy.Deploy(force)
	},
}

var bld = &cobra.Command{
	Use:   "build",
	Short: "Generate the 'site' folder.",
	Run: func(cmd *cobra.Command, args []string) {
		build.Build(force)
	},
}

func init() {
	RootCmd.AddCommand(dpl)
	RootCmd.AddCommand(bld)
	dpl.Flags().BoolVarP(&force, "force", "f", false, "force upload even if files exist")
	bld.Flags().BoolVarP(&force, "force", "f", false, "force regenerate even if files exist")
	RootCmd.Flags().BoolVarP(&force, "force", "f", false, "force regenerate and upload even if files exist")
}
