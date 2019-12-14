package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sstarcher/helm-release/git"
	"github.com/sstarcher/helm-release/helm"
	"github.com/sstarcher/helm-release/version"
)

var cfgFile string
var tag string
var tagPath string
var printComputedVersion bool
var bump string
var source string

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

		var getter version.Getter
		if source == "git" {
			source, err := git.New(dir)
			if err != nil {
				return err
			}
			getter = source.(version.Getter)
		} else if source == "helm" {
			source, err := helm.New(dir, nil)
			if err != nil {
				return err
			}
			getter = source.(version.Getter)
		} else {
			log.Fatalf("invalid input for source %s", source)
		}

		nextType := version.NewNextType(bump)
		if printComputedVersion {
			value, err := getter.Get()
			if err != nil {
				return err
			}

			ver, err := version.NextVersion(value, nextType)
			if err != nil {
				return err
			}

			_, err = os.Stdout.WriteString(ver.String())
			return err
		}

		chart, err := helm.New(dir, &tagPath)
		if err != nil {
			return err
		}

		version, err := getter.NextVersion(nextType)
		if err != nil {
			return err
		}

		log.Infof("updating the Chart.yaml to version %s", version.String())

		chart.Set(version)
		if tag != "" {
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
	rootCmd.Flags().StringVar(&bump, "bump", "", "Specifies to bump major, minor, or patch when using print-computed-version")
	rootCmd.Flags().StringVar(&source, "source", "git", "Specifies the source of the version information options (git, helm)")
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
