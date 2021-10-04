package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	// Packages
	router "github.com/mutablelogic/go-server/pkg/httprouter"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PingResponse struct {
	Version   string   `json:"version"`
	Broker    string   `json:"broker"`
	Database  string   `json:"database"`
	Retain    string   `json:"retain"`
	Connected string   `json:"connected,omitempty"`
	Count     int64    `json:"count,omitempty"`
	Topics    []string `json:"topics,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRoutePing = regexp.MustCompile(`^/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	maxResultLimit = 1000
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (p *plugin) AddHandlers(ctx context.Context, provider Provider) error {
	// Add handler for ping
	if err := provider.AddHandlerFuncEx(ctx, reRoutePing, p.ServePing); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (p *plugin) ServePing(w http.ResponseWriter, req *http.Request) {
	// Populate response
	response := PingResponse{
		Version:  p.client.Version(),
		Broker:   p.cfg.Broker,
		Database: p.cfg.Database,
		Retain:   fmt.Sprint(p.cfg.Retain),
		Topics:   p.topics.Topics(),
	}

	// Count messages in the database
	count, err := p.Count(req.Context())
	if err != nil {
		router.ServeError(w, http.StatusBadGateway, err.Error())
		return
	} else if count >= 0 {
		response.Count = count
	}

	// Set connected status
	if !p.connected.IsZero() {
		response.Connected = fmt.Sprint(time.Since(p.connected).Truncate(time.Second))
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}
