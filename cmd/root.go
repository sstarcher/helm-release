package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sstarcher/helm-release/helm"
)

var cfgFile string
var tag string
var tagPath string
var printComputedVersion bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "helm-release [CHART_PATH]",
	Short: "Determines the charts next release number",
	Long: `This plugin will use environment variables and git history to divine the next chart version.
	It will also optionally update the image tag in the values.yaml file.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		chart, err := helm.New(dir)
		if err != nil {
			return err
		}

		version, err := chart.Version()
		if err != nil {
			return err
		}

		if printComputedVersion {
			_, err = os.Stdout.WriteString(*version)
			return err
		}

		log.Infof("updating the Chart.yaml to version %s", *version)

		chart.UpdateChartVersion(*version)
		if tag != "" {
			chart.TagPath = tagPath
			err = chart.UpdateImageVersion(tag)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Info(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.helm-release.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&tag, "tag", "t", "", "Sets the docker image tag in values.yaml")
	rootCmd.Flags().StringVar(&tagPath, "path", helm.DefaultTagPath, "Sets the path to the image tag to modify in values.yaml")
	rootCmd.Flags().BoolVar(&printComputedVersion, "print-computed-version", false, "Print the computed version string to stdout")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Info(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".helm-release" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".helm-release")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}
