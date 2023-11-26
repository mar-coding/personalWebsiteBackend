package configHandler

type BaseConfig[T any] struct {
	Address        string        `yaml:"address" json:"address"`
	Domain         string        `yaml:"domain" json:"domain"`
	Origins        []string      `yaml:"origins" json:"origins"`
	Grpc           *GRPC         `yaml:"grpc" json:"grpc"`
	Rest           *REST         `yaml:"rest" json:"rest"`
	Websocket      *WEBSOCKET    `yaml:"websocket" json:"websocket"`
	GraphQL        *GRAPHQL      `yaml:"graphql" json:"graphql"`
	SocketFilePath string        `yaml:"socket_file_path" json:"socket_file_path"`
	Development    bool          `yaml:"development" json:"development"`
	GrpcClients    []*GrpcClient `yaml:"grpc_clients" json:"grpc_clients"`
	Database       *Database     `yaml:"database" json:"database"`
	Broker         *Broker       `yaml:"broker" json:"broker"`
	Logging        *Logging      `yaml:"logging" json:"logging"`
	ExtraData      T             `yaml:"extra_data" json:"extra_data"`
}

type GRPC struct {
	Port            int    `yaml:"port" json:"port"`
	CertFilePath    string `yaml:"cert_file_path" json:"cert_file_path"`
	CertKeyFilePath string `yaml:"cert_key_file_path" json:"cert_key_file_path"`
}

type REST struct {
	Port            int    `yaml:"port" json:"port"`
	CertFilePath    string `yaml:"cert_file_path" json:"cert_file_path"`
	CertKeyFilePath string `yaml:"cert_key_file_path" json:"cert_key_file_path"`
}

type WEBSOCKET struct {
	Port            int    `yaml:"port" json:"port"`
	CertFilePath    string `yaml:"cert_file_path" json:"cert_file_path"`
	CertKeyFilePath string `yaml:"cert_key_file_path" json:"cert_key_file_path"`
}

type GRAPHQL struct {
	Port int `yaml:"port" json:"port"`
}

type GrpcClient struct {
	Name           string `yaml:"name" json:"name"`
	Address        string `yaml:"address" json:"address"`
	Port           int    `yaml:"port" json:"port"`
	SocketFilePath string `yaml:"socket_file_path" json:"socket_file_path"`
	CertCAFilePath string `yaml:"cert_ca_file_path" json:"cert_ca_file_path"`
}

type Database struct {
	Mongodb  *Mongodb  `yaml:"mongodb" json:"mongodb"`   // Mongodb URI address
	MySQL    *Mysql    `yaml:"mysql" json:"mysql"`       // MySQL URI address
	Postgres *Postgres `yaml:"postgres" json:"postgres"` // Postgres URI address
	Redis    *Redis    `yaml:"redis" json:"redis"`       // Redis URI address
	Elastic  *Elastic  `yaml:"elastic" json:"elastic"`
}

type Mongodb struct {
	URI          string `yaml:"uri" json:"uri"`
	DatabaseName string `yaml:"database_name" json:"database_name"`
}

type Mysql struct {
	URI          string `yaml:"uri" json:"uri"`
	DatabaseName string `yaml:"database_name" json:"database_name"`
}

type Postgres struct {
	URI          string `yaml:"uri" json:"uri"`
	DatabaseName string `yaml:"database_name" json:"database_name"`
}

type Redis struct {
	Address  string `yaml:"address" json:"address"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Database int    `yaml:"database" json:"database"`
}

type Elastic struct {
	Addresses []string `yaml:"addresses" json:"addresses"`
	Username  string   `yaml:"username" json:"username"`
	Password  string   `yaml:"password" json:"password"`
}

type Broker struct {
	Nats     *Nats     `yaml:"nats" json:"nats"`
	RabbitMQ *RabbitMQ `yaml:"rabbitmq" json:"rabbitmq"`
}

type Nats struct {
	Address  string `yaml:"address" json:"address"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

type RabbitMQ struct {
	URI string `yaml:"uri" json:"uri"`
}

type Logging struct {
	Debug        bool   `yaml:"debug" json:"debug"`
	Handler      uint8  `yaml:"handler" json:"handler"` // Handler 0= console handler, 1= text handler, 2= json handler
	EnableCaller bool   `yaml:"enable_caller" json:"enable_caller"`
	SentryDSN    string `yaml:"sentry_dsn" json:"sentry_dsn"`
}

type LogRotation struct {
	MaxAge   int  `yaml:"max_age" json:"max_age"`
	FileSize int  `yaml:"file_size" json:"file_size"`
	Compress bool `yaml:"compress" json:"compress"`
}
