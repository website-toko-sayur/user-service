package entity

import "time"

type HealthCheck struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Uptime       string            `json:"uptime"`
	Timestamp    time.Time         `json:"timestamp"`
	Dependencies map[string]string `json:"dependencies"`
}
