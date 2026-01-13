package server

import "log"

type Config struct {
	scheme string
	host   string
	port   string
	logger *log.Logger
}

const DefaultPort = "8080"
const DefaultHost = "localhost"
const SchemeHTTP = "http"

func NewConfig(s, h, p string, logger *log.Logger) *Config {
	return &Config{s, h, p, logger}
}
func (c *Config) ServerAddr() string {
	return c.scheme + "://" + c.host + ":" + c.port
}
func (c *Config) InternalAddr() string {
	return c.host + ":" + c.port

}
