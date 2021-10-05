package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	// Package imports
	"github.com/mutablelogic/go-mosquitto/pkg/mosquitto"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MessageType string

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	messageTableName = "mqtt"
	messageIndexName = "mqtt_topic"
)

const (
	MessageTypeEmpty   = "null"
	MessageTypeText    = "text"
	MessageTypeJSON    = "json"
	MessageTypeXML     = "xml"
	MessageTypeBinary  = "byte"
	MessageTypeNumeric = "number"
	MessageTypeBoolean = "boolean"
)

var (
	// cast for first two elements of message
	messageRowCast = []reflect.Type{
		reflect.TypeOf(uint(0)),     // id
		reflect.TypeOf(time.Time{}), // ts
	}
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func HasSchema(p pool, v string) error {
	// Get a connection
	conn := p.Get()
	if conn == nil {
		return ErrInternalAppError.With("Missing database connection")
	}
	defer p.Put(conn)

	// Find the schema
	schemas := conn.Schemas()
	for _, schema := range schemas {
		if schema == v {
			return nil
		}
	}

	// Return not found
	return ErrNotFound.Withf("Schema not found: %q", v)
}

func (p *plugin) AddSchema(ctx context.Context) error {
	// Get a connection
	conn := p.Get()
	if conn == nil {
		return ErrInternalAppError.With("Missing database connection")
	}
	defer p.Put(conn)

	return conn.Do(ctx, 0, func(txn SQTransaction) error {
		// Create the table
		if _, err := txn.Query(N(messageTableName).WithSchema(p.cfg.Database).CreateTable(
			C("id").WithAutoIncrement().WithType("INTEGER"),
			C("ts").NotNull(),
			C("topic").NotNull(),
			C("type"),
			C("payload").WithType("BLOB"),
		).IfNotExists()); err != nil {
			return err
		}
		// Create the index on topic
		if _, err := txn.Query(N(messageIndexName).WithSchema(p.cfg.Database).CreateIndex(
			messageTableName, "topic",
		).IfNotExists()); err != nil {
			return err
		}
		// Return success
		return nil
	})
}

// Add a messge to the database
func (p *plugin) AddMessage(ctx context.Context, msg *mosquitto.Event) error {
	// Get a connection
	conn := p.Get()
	if conn == nil {
		return ErrInternalAppError.With("Missing database connection")
	}
	defer p.Put(conn)

	// Insert the data in a transaction
	return conn.Do(ctx, 0, func(txn SQTransaction) error {
		t := toType(msg.Data)
		if _, err := txn.Query(N(messageTableName).WithSchema(p.cfg.Database).Insert(
			"ts", "topic", "type", "payload",
		), time.Now(), msg.Topic, string(t), msg.Data); err != nil {
			return err
		}

		// Return success
		return nil
	})
}

// Return message count
func (p *plugin) Count(ctx context.Context) (int64, error) {
	// Get a connection
	conn := p.Get()
	if conn == nil {
		return 0, ErrInternalAppError.With("Missing database connection")
	}
	defer p.Put(conn)

	// Count within a transaction
	n := int64(0)
	if err := conn.Do(ctx, 0, func(txn SQTransaction) error {
		n = txn.Count(p.cfg.Database, messageTableName)
		return nil
	}); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

// Delete messages older than the retain cycle
func (p *plugin) RetainCycle(ctx context.Context) (int, error) {
	// Get a connection
	conn := p.Get()
	if conn == nil {
		return 0, ErrInternalAppError.With("Missing database connection")
	}
	defer p.Put(conn)

	// Delete older messages in a transaction
	var n int
	if err := conn.Do(ctx, 0, func(txn SQTransaction) error {
		retain := p.cfg.Retain.Seconds()
		if r, err := txn.Query(N(messageTableName).WithSchema(p.cfg.Database).Delete(
			Q("CAST((JulianDay('now') - JulianDay(ts)) * 24 * 3600 AS INTEGER) > ?"),
		), retain); err != nil {
			return err
		} else {
			n = r.RowsAffected()
		}
		// Return success
		return nil
	}); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

// Query for messages by type and order results
func (p *plugin) Query(ctx context.Context, t string, order string, limit uint) (SQResults, error) {
	var results SQResults

	// Get a connection
	conn := p.Get()
	if conn == nil {
		return nil, ErrInternalAppError.With("Missing database connection")
	}
	defer p.Put(conn)

	// Perform the transaction
	if err := conn.Do(ctx, 0, func(txn SQTransaction) error {
		var err error
		var params []interface{}
		// Create the select statement
		s := S(N(messageTableName).WithSchema(p.cfg.Database)).
			To(N("id"), N("ts"), N("topic"), N("type"), N("payload"))
		// Append parameters
		if limit > 0 {
			s = s.WithLimitOffset(limit, 0)
		}
		if t != "" {
			s = s.Where(Q("type = ?"))
			params = append(params, t)
		}
		for _, order := range strings.FieldsFunc(order, func(c rune) bool {
			return c == ',' || c == ' '
		}) {
			var desc bool
			order = strings.TrimSpace(order)
			if strings.HasPrefix(order, "-") {
				desc = true
			}
			order = strings.TrimLeftFunc(order, func(c rune) bool {
				return c == '-' || c == '+'
			})
			if desc {
				s = s.Order(N(order).WithDesc())
			} else {
				s = s.Order(N(order))
			}
		}
		// Run the query and return any errors
		results, err = txn.Query(s, params...)
		return err
	}); err != nil {
		return nil, err
	} else {
		return results, nil
	}
}

// Get message by id for messages by type and order results
func (p *plugin) GetMessage(ctx context.Context, id int64) (SQResults, error) {
	var results SQResults

	// Get a connection
	conn := p.Get()
	if conn == nil {
		return nil, ErrInternalAppError.With("Missing database connection")
	}
	defer p.Put(conn)

	// Perform the transaction
	if err := conn.Do(ctx, 0, func(txn SQTransaction) error {
		var err error

		// Create the select statement
		s := S(N(messageTableName).WithSchema(p.cfg.Database)).
			To(N("id"), N("ts"), N("topic"), N("type"), N("payload")).
			Where(Q("id = ?"))
		// Run query and return results
		results, err = txn.Query(s, id)
		return err
	}); err != nil {
		return nil, err
	} else {
		return results, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE  METHODS

func toType(data []byte) MessageType {
	// Check for empty
	if len(data) == 0 {
		return MessageTypeEmpty
	}
	// Check for number
	if isNumber(data) {
		return MessageTypeNumeric
	}
	// Check for boolean
	if isBoolean(data) {
		return MessageTypeBoolean
	}
	// Check for JSON structs
	if isJSON(data) {
		return MessageTypeJSON
	}
	// Check for XML
	if isXML(data) {
		return MessageTypeXML
	}
	// Check for UTF-8 string
	if utf8.Valid(data) {
		return MessageTypeText
	}
	// Default to binary
	return MessageTypeBinary
}

func isNumber(data []byte) bool {
	data = bytes.TrimSpace(data)
	if _, err := strconv.ParseFloat(string(data), 64); err == nil {
		return true
	} else {
		return false
	}
}

func isBoolean(data []byte) bool {
	data = bytes.TrimSpace(data)
	if _, err := strconv.ParseBool(string(data)); err == nil {
		return true
	} else {
		return false
	}
}

func isXML(data []byte) bool {
	data = bytes.TrimSpace(data)

	// Sanity check data to ensure < and > at beginning and end
	if !bytes.HasPrefix(data, []byte("<")) || !bytes.HasSuffix(data, []byte(">")) {
		return false
	}

	// Now go the long way around
	decoder := xml.NewDecoder(bytes.NewBuffer(data))
	for {
		err := decoder.Decode(new(interface{}))
		if err != nil {
			return err == io.EOF
		}
	}
}

func isJSON(data []byte) bool {
	data = bytes.TrimSpace(data)

	// Sanity check data to ensure { or [ at beginning
	if !bytes.HasPrefix(data, []byte("{")) && !bytes.HasPrefix(data, []byte("[")) {
		return false
	}
	// Sanity check data to ensure } or ] at end
	if !bytes.HasSuffix(data, []byte("}")) && !bytes.HasSuffix(data, []byte("]")) {
		return false
	}

	// Now go the long way around
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
