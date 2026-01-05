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
