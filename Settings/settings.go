package Settings

type QDBSettings struct {
	RootPath        string
	WatchdgoSec     int
	ActiveToDelayed int
	DelayedToFailed int
}

type RedConSettings struct {
	Port     int
	Password string
}

type HttpServer struct {
	Port     int
	Username string
	Password string
}

type GoQSettings struct {
	SettingsDBPath string
}
