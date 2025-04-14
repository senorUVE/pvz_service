package repository

type DBConfig struct {
	DBDriver string `mapstructure:"db_driver"`
	DBUser   string `mapstructure:"db_user"`
	DBPass   string `mapstructure:"db_password"`
	DBHost   string `mapstructure:"db_host"`
	DBPort   string `mapstructure:"db_port"`
	DBName   string `mapstructure:"db_name"`
	DBSSL    string `mapstructure:"db_ssl"`
}
