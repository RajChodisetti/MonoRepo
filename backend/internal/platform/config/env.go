package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type envParser struct {
	errs []error
}

func (p *envParser) join() error {
	return joinErrors(p.errs)
}

func (p *envParser) string(key, fallback string) string {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	return value
}

func (p *envParser) csv(key string, fallback []string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	if len(values) == 0 {
		p.errs = append(p.errs, fmt.Errorf("%s must contain at least one value", key))
		return fallback
	}
	return values
}

func (p *envParser) int(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		p.errs = append(p.errs, fmt.Errorf("%s must be a valid integer", key))
		return fallback
	}
	return value
}

func (p *envParser) bool(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}

	value, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		p.errs = append(p.errs, fmt.Errorf("%s must be true or false", key))
		return fallback
	}
	return value
}

func (p *envParser) duration(key string, fallback time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}

	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		p.errs = append(p.errs, fmt.Errorf("%s must be a valid duration (example: 30s, 5m, 1h)", key))
		return fallback
	}
	return value
}

func (p *envParser) location(key, fallback string) *time.Location {
	name := p.string(key, fallback)
	loc, err := time.LoadLocation(name)
	if err != nil {
		p.errs = append(p.errs, fmt.Errorf("%s must be a valid IANA timezone", key))
		fallbackLoc, fallbackErr := time.LoadLocation(fallback)
		if fallbackErr != nil {
			return time.UTC
		}
		return fallbackLoc
	}
	return loc
}

func (p *envParser) listenAddr() string {
	if addr := strings.TrimSpace(os.Getenv("HTTP_ADDR")); addr != "" {
		return normalizeListenAddr(addr)
	}

	portKey := ""
	portValue := ""
	switch {
	case isSet("HTTP_PORT"):
		portKey, portValue = "HTTP_PORT", strings.TrimSpace(os.Getenv("HTTP_PORT"))
	case isSet("PORT"):
		portKey, portValue = "PORT", strings.TrimSpace(os.Getenv("PORT"))
	default:
		return ":8080"
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		p.errs = append(p.errs, fmt.Errorf("%s must be a valid TCP port between 1 and 65535", portKey))
		return ":8080"
	}
	return fmt.Sprintf(":%d", port)
}

func isSet(key string) bool {
	raw, ok := os.LookupEnv(key)
	return ok && strings.TrimSpace(raw) != ""
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
