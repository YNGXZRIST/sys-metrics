// Package internal holds unexported configuration types for the db package.
package internal

// Config holds the PostgreSQL connection DSN.
type Config struct {
	DNS string
}
