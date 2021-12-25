package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thomas-maurice/netlab/pkg/version"
)

var (
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "netlab",
	Short: "Create and tear down network inrastructures easily",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	InitRootCmd()
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version number",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Git Hash: %s\nBuild Host: %s\nBuild Time: %s\nBuild Tag: %s\n", version.BuildHash, version.BuildHost, version.BuildTime, version.Version)
	},
}

func InitRootCmd() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "network.yaml", "Configuration file")

	InitUpCmd()
	InitDownCmd()
	InitDHCPUpCmd()

	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(UpCmd)
	rootCmd.AddCommand(DownCmd)
	rootCmd.AddCommand(DHCPUpCmd)
	rootCmd.AddCommand(DHCPDownCmd)
}
