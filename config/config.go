package config

import "github.com/BurntSushi/toml"

type Config struct {
	HttpAddr string `toml:"http_addr"`
	NsqdAddr string `toml:"nsqd_addr"`
	Db DbConfig
	Redis Redis
}


type DbConfig struct {
	Addr string `toml:"addr"`
	User string `toml:"user"`
	Pass string `toml:"password"`
	Database string `toml:"database"`
}

type Redis struct {
	Addr string `toml:"addr"`
	Db 	 int `toml:"db"`
	Pass string `toml:"password"`
}

func DefaultDbcfg() DbConfig {
	return DbConfig{
		Addr:"127.0.0.1:3306",
		User:"root",
		Pass:"123456",
		Database:"imdata",
	}
}

func DefaultRedis() Redis  {
	return Redis{
		Addr:"127.0.0.1:6379",
		Db:1,
		Pass:"",
	}
}


var DefaultConf = Config{
	HttpAddr:"localhost:8080",
	NsqdAddr:"localhost:9876",
	Db:DefaultDbcfg(),
	Redis:DefaultRedis(),
}

func LoadFile(filename string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filename, &config); err != nil {
		return &DefaultConf, err
	}
	return &config, nil

}