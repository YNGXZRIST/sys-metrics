package config

import (
	"fmt"
	"slices"
	"strings"
	"sys-metrics/internal/common"
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
func ValidateMode(mode string) error {
	validModes := []string{common.TypeModeDevelopment, common.TypeModeProduction}
	containsTen := slices.Contains(validModes, mode)
	if !containsTen {
		return fmt.Errorf("invalid mode: %s,valide types: %v", mode, validModes)
	}
	return nil
}
