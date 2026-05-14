// Package config provides shared helpers for parsing server addresses and validating run mode.
package config

import (
	"fmt"
	"slices"
	"strings"
	"sys-metrics/internal/common"
)

// generate:reset

// ServerAddress is a parsed host:port pair.
type ServerAddress struct {
	Full string
	Host string
	Port string
}

// HostPortSetter applies parsed host and port to flags or option structs.
type HostPortSetter interface {
	SetHostPort(host, port string)
}

// ParseServerAddress parses a string of the form "host:port".
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

// ParseAndSetHostPort parses address and calls setter.SetHostPort.
func ParseAndSetHostPort(address string, setter HostPortSetter) error {
	addr, err := ParseServerAddress(address)
	if err != nil {
		return fmt.Errorf("parse address error: %w", err)
	}
	setter.SetHostPort(addr.Host, addr.Port)
	return nil
}

// ValidateMode ensures mode is development or production.
func ValidateMode(mode string) error {
	validModes := []string{common.TypeModeDevelopment, common.TypeModeProduction}
	containsTen := slices.Contains(validModes, mode)
	if !containsTen {
		return fmt.Errorf("invalid mode: %s,valid types: %v", mode, validModes)
	}
	return nil
}
