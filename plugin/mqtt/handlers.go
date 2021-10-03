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
	Version   string `json:"version"`
	Broker    string `json:"broker"`
	Database  string `json:"database"`
	Retain    string `json:"retain"`
	Connected string `json:"connected"`
	Count     int64  `json:"count"`
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
	// Count messages in the database
	count, err := p.Count(req.Context())
	if err != nil {
		router.ServeError(w, http.StatusBadGateway, err.Error())
		return
	}

	// Populate response
	response := PingResponse{
		Version:  p.client.Version(),
		Broker:   p.cfg.Broker,
		Database: p.cfg.Database,
		Retain:   fmt.Sprint(p.cfg.Retain),
		Count:    count,
	}
	if !p.connected.IsZero() {
		response.Connected = fmt.Sprint(time.Since(p.connected).Truncate(time.Second))
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}
