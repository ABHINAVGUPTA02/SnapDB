package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ABHINAVGUPTA02/SnapDB/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage SnapDB configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the SnapDB config directory and config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureDirs(); err != nil {
			return err
		}
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}
		fmt.Printf("Initialized SnapDB config at %s\n", path)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show SnapDB config with secrets redacted",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var configSetFlags struct {
	name        string
	defaultSet  bool
	dbType      string
	user        string
	passwordEnv string
	host        string
	port        string
	dbname      string
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create or update a database profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if configSetFlags.name == "" {
			return fmt.Errorf("--name is required")
		}
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		profile := cfg.Profiles[configSetFlags.name]
		if configSetFlags.dbType != "" {
			profile.Type = configSetFlags.dbType
		}
		if configSetFlags.user != "" {
			profile.Username = configSetFlags.user
		}
		if configSetFlags.passwordEnv != "" {
			profile.PasswordEnv = configSetFlags.passwordEnv
		}
		if configSetFlags.host != "" {
			profile.Host = configSetFlags.host
		}
		if configSetFlags.port != "" {
			profile.Port = configSetFlags.port
		}
		if configSetFlags.dbname != "" {
			profile.Database = configSetFlags.dbname
		}
		if profile.Host == "" {
			profile.Host = "localhost"
		}
		if profile.Port == "" {
			profile.Port = defaultPort(profile.Type)
		}

		cfg.Profiles[configSetFlags.name] = profile
		if configSetFlags.defaultSet || cfg.DefaultProfile == "" {
			cfg.DefaultProfile = configSetFlags.name
		}
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Saved profile %s\n", configSetFlags.name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd, configShowCmd, configSetCmd)

	configSetCmd.Flags().StringVar(&configSetFlags.name, "name", "", "Profile name")
	configSetCmd.Flags().BoolVar(&configSetFlags.defaultSet, "default", false, "Set as default profile")
	configSetCmd.Flags().StringVar(&configSetFlags.dbType, "type", "", "Database type")
	configSetCmd.Flags().StringVar(&configSetFlags.user, "user", "", "Database user")
	configSetCmd.Flags().StringVar(&configSetFlags.passwordEnv, "password-env", "", "Environment variable containing the database password")
	configSetCmd.Flags().StringVar(&configSetFlags.host, "host", "", "Database host")
	configSetCmd.Flags().StringVar(&configSetFlags.port, "port", "", "Database port")
	configSetCmd.Flags().StringVar(&configSetFlags.dbname, "dbname", "", "Database name")
}
