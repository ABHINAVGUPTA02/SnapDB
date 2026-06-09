package cmd

import (
	"fmt"
	"os"

	"github.com/ABHINAVGUPTA02/SnapDB/config"
	"github.com/ABHINAVGUPTA02/SnapDB/internal/ops"
	"github.com/spf13/cobra"
)

type dbFlags struct {
	profile     string
	dbType      string
	user        string
	password    string
	passwordEnv string
	host        string
	port        string
	dbname      string
}

func addDatabaseFlags(command *cobra.Command, flags *dbFlags) {
	command.Flags().StringVar(&flags.profile, "profile", "", "Config profile to use")
	command.Flags().StringVar(&flags.dbType, "type", "", "Database type: mysql, postgres, mongodb, redis")
	command.Flags().StringVar(&flags.user, "user", "", "Database user")
	command.Flags().StringVar(&flags.password, "password", "", "Database password for this run")
	command.Flags().StringVar(&flags.passwordEnv, "password-env", "", "Environment variable containing the database password")
	command.Flags().StringVar(&flags.host, "host", "", "Database host")
	command.Flags().StringVar(&flags.port, "port", "", "Database port")
	command.Flags().StringVar(&flags.dbname, "dbname", "", "Database name")
}

func resolveTarget(flags dbFlags, positionalType string) (ops.DatabaseTarget, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return ops.DatabaseTarget{}, err
	}

	profileName := flags.profile
	if profileName == "" {
		profileName = cfg.DefaultProfile
	}

	var target ops.DatabaseTarget
	if profileName != "" {
		profile, ok := cfg.Profiles[profileName]
		if !ok {
			return ops.DatabaseTarget{}, fmt.Errorf("profile %q not found", profileName)
		}
		target = ops.TargetFromProfile(profileName, profile)
	}

	if positionalType != "" {
		target.Type = positionalType
	}
	if flags.dbType != "" {
		target.Type = flags.dbType
	}
	if flags.user != "" {
		target.User = flags.user
	}
	if flags.password != "" {
		target.Password = flags.password
	}
	if flags.passwordEnv != "" {
		target.PasswordEnv = flags.passwordEnv
	}
	if flags.host != "" {
		target.Host = flags.host
	}
	if flags.port != "" {
		target.Port = flags.port
	}
	if flags.dbname != "" {
		target.Database = flags.dbname
	}

	if target.Password == "" && target.PasswordEnv != "" {
		target.Password = os.Getenv(target.PasswordEnv)
	}
	if target.Port == "" {
		target.Port = defaultPort(target.Type)
	}
	if target.Host == "" {
		target.Host = "localhost"
	}

	return target, nil
}

func defaultPort(databaseType string) string {
	switch databaseType {
	case "mysql", "mariadb":
		return "3306"
	case "postgres", "postgresql":
		return "5432"
	case "mongodb", "mongo":
		return "27017"
	case "redis":
		return "6379"
	default:
		return ""
	}
}
