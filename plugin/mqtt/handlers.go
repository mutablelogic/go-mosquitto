package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	// Packages
	router "github.com/mutablelogic/go-server/pkg/httprouter"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TopicRequest struct {
	Topic string `json:"topic"`
}

type MessageRequest struct {
	Topic string `json:"topic"`
	Type  string `json:"type"`
	Limit uint   `json:"limit"`
	Order string `json:"order"`
}

type MessageResponse struct {
	Id        uint        `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"ts"`
	Topic     string      `json:"topic"`
	Payload   interface{} `json:"payload,omitempty"`
	Value     interface{} `json:"value,omitempty"`
}

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
	reRoutePing     = regexp.MustCompile(`^/?$`)
	reRouteTopics   = regexp.MustCompile(`^/t/?$`)
	reRouteMessages = regexp.MustCompile(`^/m/?$`)
	reRouteMessage  = regexp.MustCompile(`^/m/(\d+)/?$`)
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
	// Add handler for topics
	if err := provider.AddHandlerFuncEx(ctx, reRouteTopics, p.ServeTopicList); err != nil {
		return err
	}
	if err := provider.AddHandlerFuncEx(ctx, reRouteTopics, p.ServeTopicSubscribe, http.MethodPut, http.MethodPost, http.MethodDelete); err != nil {
		return err
	}
	// Add handler for messages
	if err := provider.AddHandlerFuncEx(ctx, reRouteMessages, p.ServeMessageList); err != nil {
		return err
	}
	if err := provider.AddHandlerFuncEx(ctx, reRouteMessage, p.ServeMessage); err != nil {
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

func (p *plugin) ServeTopicList(w http.ResponseWriter, req *http.Request) {
	// Serve response
	router.ServeJSON(w, p.topics.Topics(), http.StatusOK, 2)
}

func (p *plugin) ServeTopicSubscribe(w http.ResponseWriter, req *http.Request) {
	// Get topic
	var topic TopicRequest
	if err := router.RequestBody(req, &topic); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	} else if topic.Topic == "" {
		router.ServeError(w, http.StatusBadRequest, "topic is required")
		return
	}

	switch req.Method {
	case http.MethodDelete:
		// Unsubscribe
		if err := p.Unubscribe(topic.Topic); err != nil {
			router.ServeError(w, http.StatusBadGateway, err.Error())
			return
		}
	default:
		// Subscribe
		if err := p.Subscribe(topic.Topic); err != nil {
			router.ServeError(w, http.StatusBadGateway, err.Error())
			return
		}
	}

	// Serve response
	router.ServeJSON(w, p.topics.Topics(), http.StatusOK, 2)
}

func (p *plugin) ServeMessageList(w http.ResponseWriter, req *http.Request) {
	// Get message request parameters
	var q MessageRequest
	if err := router.RequestQuery(req, &q); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check query parameters
	q.Limit = uintMin(maxResultLimit, q.Limit)
	if q.Order == "" {
		// Default to ordering by descending timestamp
		q.Order = "-ts"
	}

	// Run the query
	results, err := p.Query(req.Context(), q.Type, q.Order, q.Limit)
	if err != nil {
		router.ServeError(w, http.StatusBadGateway, err.Error())
		return
	}
	//	defer results.Close()

	// Serve response
	router.ServeJSON(w, makeResponse(results, int(q.Limit)), http.StatusOK, 2)
}

func (p *plugin) ServeMessage(w http.ResponseWriter, req *http.Request) {
	// Get message request parameters
	params := router.RequestParams(req)
	id, err := strconv.ParseInt(params[0], 0, 64)
	if err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Run the query
	results, err := p.GetMessage(req.Context(), id)
	if err != nil {
		router.ServeError(w, http.StatusBadGateway, err.Error())
		return
	}
	//	defer results.(*sqlite3.Results).Close()

	// Serve response
	response := makeResponse(results, 1)
	if len(response) > 0 {
		router.ServeJSON(w, response[0], http.StatusOK, 2)
	} else {
		router.ServeError(w, http.StatusNotFound)
	}
}

func makeResponse(r SQResults, cap int) []MessageResponse {
	// The results are id, ts, topic, type and payload
	result := make([]MessageResponse, 0, cap)
	for {
		row := r.Next(messageRowCast...)
		if row == nil {
			break
		}
		message := MessageResponse{
			Id:        row[0].(uint),
			Timestamp: row[1].(time.Time),
			Topic:     row[2].(string),
			Type:      row[3].(string),
		}
		str := strings.TrimSpace(string(row[4].([]byte)))
		switch message.Type {
		case MessageTypeText:
			message.Payload = str
			message.Value = str
		case MessageTypeNumeric:
			message.Payload = str
			if n, err := strconv.ParseFloat(str, 64); err == nil {
				message.Value = n
			}
		case MessageTypeBoolean:
			message.Payload = str
			if n, err := strconv.ParseBool(str); err == nil {
				message.Value = n
			}
		case MessageTypeBinary:
			message.Payload = row[4]
		case MessageTypeXML:
			message.Payload = str
		case MessageTypeJSON:
			message.Payload = str
			if err := json.Unmarshal(bytes.TrimSpace(row[4].([]byte)), &message.Value); err != nil {
				fmt.Println(message.Id, err, string(row[4].([]byte)))
			}
		}
		result = append(result, message)
	}
	return result
}
