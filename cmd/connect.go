package cmd

import (
	"context"
	"fmt"

	"github.com/ABHINAVGUPTA02/SnapDB/internal/db"
	"github.com/spf13/cobra"
)

var (
	dbType   string
	user     string
	password string
	host     string
	port     string
	dbname   string
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to SnapDB",
	Long:  `Establishes a connection to a supported database (MySQL, Postgres, MongoDB, Redis, etc.).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Connecting to database %s at %s:%s\n", dbType, host, port)

		if host == "" || port == "" || dbname == "" || user == "" || password == "" {
			return fmt.Errorf("type, user, password, host, port, and dbname are required")
		}

		if err := connectToDB(dbType, user, password, host, port, dbname); err != nil {
			return err
		}

		saveConnection()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().StringVar(&dbType, "type", "mysql", "Database type")
	connectCmd.Flags().StringVar(&user, "user", "root", "Database user")
	connectCmd.Flags().StringVar(&password, "password", "", "Database password")
	connectCmd.Flags().StringVar(&host, "host", "localhost", "Database host")
	connectCmd.Flags().StringVar(&port, "port", "3306", "Database port")
	connectCmd.Flags().StringVar(&dbname, "dbname", "", "Database name")

	connectCmd.MarkFlagRequired("type")
	connectCmd.MarkFlagRequired("user")
	connectCmd.MarkFlagRequired("password")
	connectCmd.MarkFlagRequired("host")
	connectCmd.MarkFlagRequired("port")
	connectCmd.MarkFlagRequired("dbname")
}

func connectToDB(dbType string, user string, password string, host string, port string, dbname string) error {
	switch dbType {
	case "mysql":
		connector := &db.MySQLConnector{}
		conn, err := connector.Connect(user, password, host, port, dbname)
		if err != nil {
			return fmt.Errorf("failed to connect to the database: %w", err)
		}
		defer conn.Close()
		fmt.Printf("Successfully connected to %s database %s at %s:%s\n", dbType, dbname, host, port)

	case "postgres":
		connector := &db.PostgresqlConnector{}
		conn, err := connector.Connect(user, password, host, port, dbname)
		if err != nil {
			return fmt.Errorf("failed to connect to the database: %w", err)
		}
		defer conn.Close()
		fmt.Printf("Successfully connected to %s database %s at %s:%s\n", dbType, dbname, host, port)

	case "mongo":
		connector := &db.MongoConnector{}
		conn, err := connector.Connect(user, password, host, port, dbname)
		if err != nil {
			return fmt.Errorf("failed to connect to the database: %w", err)
		}
		defer conn.Disconnect(context.Background())
		fmt.Printf("Successfully connected to %s database %s at %s:%s\n", dbType, dbname, host, port)

	default:
		return fmt.Errorf("unknown database type: %s", dbType)
	}
	return nil
}

func saveConnection() {
	fmt.Printf("Connection verified. Use `snapdb config set` to save a profile without storing plaintext credentials.\n")
}
