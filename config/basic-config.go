package config

// DBConfig is the configurations for connecting database
type DBConfig struct {
	Driver string // db driver name: sqlite, mysql, postgres
	DSN    string // db connection string
}

// HTTPConfig is the configurations for HTTP server
type HTTPConfig struct {
	Addr string // listen address: ":8080"
	//Https       bool   `json:"https"`    // enable https?
	//TLSCertPath string `json:"tls_cert_path"` // path to tls cert file
	//TLSKeyPath  string `json:"tls_key_path"`  // path to tls key file
}

// BaseConfig includes common config for services
type BaseConfig struct {
	DB       DBConfig   // database config
	HTTP     HTTPConfig // http listen config
	LogLevel string     // log level
}
