package config

// Default path configurations and global settings for Envku CLI
var (
	AppName    = "envku"
	AppVersion = "1.0.0"
)

const (
	DefaultConfigDir = "/etc/envku"
	ServiceFile    = "/etc/systemd/system/envku.service"
	NginxSitesAvail  = "/etc/nginx/sites-available"
	NginxSitesEnable = "/etc/nginx/sites-enabled"
)
