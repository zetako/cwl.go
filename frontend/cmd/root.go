package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cwl.go",
	Short: "cwl executor written in golang",
	Long:  `cwl.go can parse a group of cwl document and execute it. It also extend some custom feature of cwl.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cwl.go.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&overallFeatureSwitch, "disable-plus-loop-set", "L", false, "Disable +loop feature set")
	rootCmd.PersistentFlags().BoolVar(&flags.DisableLoopFeature, "disable-loop-feature", false, "Disable loop feature")
	rootCmd.PersistentFlags().BoolVar(&flags.DisableLastNonNull, "disable-last-non-null", false, "Disable last_non_null pickValue method")
	rootCmd.PersistentFlags().IntVar(&flags.MaxLoopCount, "max-loop-count", 0, "Max loop iter count allowed")

	rootCmd.PersistentFlags().IntVar(&flags.MaxWorkflowNested, "max-nested", 10, "Max nested sub workflow count allowed")
	rootCmd.PersistentFlags().IntVar(&flags.MaxParallelLimit, "max-parallel", 0, "Max step task run in parallel")
	rootCmd.PersistentFlags().IntVar(&flags.MaxScatterLimit, "max-scatter", 0, "Max scatter task run in parallel")

	rootCmd.PersistentFlags().DurationVar(&flags.TotalTimeLimit, "timeout", 0, "Timeout for entire run")
	rootCmd.PersistentFlags().DurationVar(&flags.StepTimeLimit, "step-timeout", 0, "timeout for single step")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
