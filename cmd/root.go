package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/superfly/flyctl/flyctl"
	"github.com/superfly/flyctl/internal/client"
)

var ErrAbort = errors.New("abort")
var flyctlClient *client.Client

var rootCmd = &Command{
	Command: &cobra.Command{
		Use:   "flyctl",
		Short: "The Fly CLI",
		Long:  `flyctl is a command line interface for the Fly.io platform`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			flyctlClient = client.NewClient()
		},
	},
}

// GetRootCommand - root for commands
func GetRootCommand() *cobra.Command {
	return rootCmd.Command
}

// Execute - root command execution
func Execute() {
	defer flyctl.BackgroundTaskWG.Wait()

	err := rootCmd.Execute()
	checkErr(err)
}

func init() {
	rootCmd.PersistentFlags().StringP("access-token", "t", "", "Fly API Access Token")
	viper.BindPFlag(flyctl.ConfigAPIToken, rootCmd.PersistentFlags().Lookup("access-token"))

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	viper.BindPFlag(flyctl.ConfigVerboseOutput, rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.AddCommand(
		newAuthCommand(),
		newAppStatusCommand(),
		newAppListCommand(),
		newAppReleasesListCommand(),
		newAppLogsCommand(),
		newAppSecretsCommand(),
		newVersionCommand(),
		newDeployCommand(),
		newAppInfoCommand(),
		newBuildsCommand(),
		newAppHistoryCommand(),
		newCertificatesCommand(),
		newDocsCommand(),
		newIPAddressesCommand(),
		newConfigCommand(),
		newScaleCommand(),
		newPlatformCommand(),
		newCurlCommand(),
		newRegionsCommand(),
		newOpenCommand(),
	)

	initConfig()
}

func initConfig() {
	flyctl.InitConfig()
	flyctl.CheckForUpdate()
}

func checkErr(err error) {
	if err == nil {
		return
	}

	if !isCancelledError(err) {
		fmt.Println(aurora.Red("Error"), err)
	}

	safeExit()
}

func isCancelledError(err error) bool {
	if err == ErrAbort {
		return true
	}

	if err == context.Canceled {
		return true
	}

	if merr, ok := err.(*multierror.Error); ok {
		if len(merr.Errors) == 1 && merr.Errors[0] == context.Canceled {
			return true
		}
	}

	return false
}

func safeExit() {
	flyctl.BackgroundTaskWG.Wait()

	os.Exit(1)
}
