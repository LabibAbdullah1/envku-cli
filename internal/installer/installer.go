package installer

// Component defines the type of installable software component
type Component string

const (
	ComponentNginx   Component = "nginx"
	ComponentMySQL   Component = "mysql"
	ComponentRedis   Component = "redis"
	ComponentNodeNVM Component = "node-nvm"
)

// SupportedPHPVersions defines the modern supported PHP versions
var SupportedPHPVersions = []string{"8.2", "8.3", "8.4", "8.5"}

// PackageStatus represents the installation and service state of a component
type PackageStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Active    bool   `json:"active"`
	Version   string `json:"version,omitempty"`
}

// Installer manages component installations and status queries
type Installer struct {
	Nginx *NginxInstaller
	PHP   *PHPInstaller
	MySQL *MySQLInstaller
	Redis *RedisInstaller
	Node  *NodeInstaller
}

// New creates and initializes an Installer instance with all sub-installers
func New() *Installer {
	return &Installer{
		Nginx: NewNginxInstaller(),
		PHP:   NewPHPInstaller(),
		MySQL: NewMySQLInstaller(),
		Redis: NewRedisInstaller(),
		Node:  NewNodeInstaller(),
	}
}
