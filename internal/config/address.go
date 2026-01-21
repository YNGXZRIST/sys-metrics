package config

import (
	"fmt"
	"strings"
)

type ServerAddress struct {
	Full string
	Host string
	Port string
}

type HostPortSetter interface {
	SetHostPort(host, port string)
}

func ParseServerAddress(address string) (*ServerAddress, error) {
	split := strings.SplitN(address, ":", 2)
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid server address format: %s", address)
	}
	return &ServerAddress{
		Full: address,
		Host: split[0],
		Port: split[1],
	}, nil
}

func ParseAndSetHostPort(address string, setter HostPortSetter) error {
	addr, err := ParseServerAddress(address)
	if err != nil {
		return err
	}
	setter.SetHostPort(addr.Host, addr.Port)
	return nil
}
