package ops

import "github.com/ABHINAVGUPTA02/SnapDB/config"

type DatabaseTarget struct {
	Type        string
	Host        string
	Port        string
	User        string
	Password    string
	PasswordEnv string
	Database    string
	Profile     string
}

func TargetFromProfile(name string, profile config.Profile) DatabaseTarget {
	return DatabaseTarget{
		Type:        profile.Type,
		Host:        profile.Host,
		Port:        profile.Port,
		User:        profile.Username,
		PasswordEnv: profile.PasswordEnv,
		Database:    profile.Database,
		Profile:     name,
	}
}
