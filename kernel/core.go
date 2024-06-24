package kernel

type dbConfig struct {
	Host     string
	User     string
	Port     string
	Password string
	Database string
	TimeZone string
}

type Config struct {
	DB dbConfig
}

type core struct {
	Config Config

	AppName string
}

var Kernel *core

func NewKernel(appName string) *core {
	return &core{Config: Config{}, AppName: appName}
}
