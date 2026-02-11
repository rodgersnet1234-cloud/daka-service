package utils

import (
	"log"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Name string `yaml:"name"`
		ID   string `yaml:"id"`
	} `yaml:"server"`
	Consul struct {
		Port    int    `yaml:"port"`
		Server  string `yaml:"server"`
		LocalIp bool   `yaml:"localIp"`
	} `yaml:"consul"`
	Nats struct {
		Port string `yaml:"port"`
		Ip   string `yaml:"ip"`
	} `yaml:"nats"`
	Config struct {
		GateWayUrl string `yaml:"gateway_url"`
		Token      string `yaml:"token"`
	} `yaml:"config"`
	Mysql struct {
		Port     int    `yaml:"port"`
		Ip       string `yaml:"ip"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"mysql"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}
