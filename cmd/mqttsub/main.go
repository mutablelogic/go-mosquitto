package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	// Packages
	"github.com/mutablelogic/go-mosquitto/pkg/app"
	"github.com/mutablelogic/go-mosquitto/pkg/config"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	flagHost    = flag.String("host", "test.mosquitto.org", "MQTT broker host")
	flagQos     = flag.Int("qos", 0, "MQTT QoS")
	flagVersion = flag.Bool("version", false, "Print version")
	flagTimeout = flag.Duration("timeout", 10*time.Second, "Connection Timeout")
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <flags> [topic] [topic]...\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output(), "\nFlags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Output version and bomb out
	if *flagVersion {
		config.PrintVersion(flag.CommandLine.Output())
		os.Exit(0)
	}

	// Topics to subscribe to
	topics := flag.Args()
	if len(topics) == 0 {
		topics = []string{"#"}
	}

	// Create a context which cancels on CTRL+C
	ctx := HandleSignal()

	// Connect with timeout
	fmt.Printf("Connecting to %q with timeout %v\n", *flagHost, *flagTimeout)
	connectctx, cancel := context.WithTimeout(ctx, *flagTimeout)
	defer cancel()
	app, err := app.NewApp(connectctx, *flagHost, *flagQos)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	fmt.Println("Press CTRL+C to end")
	if err := app.Run(ctx, topics...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func HandleSignal() context.Context {
	// Handle signals - call cancel when interrupt received
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}
