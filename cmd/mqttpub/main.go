package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <flags> [topic] [data]...\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output(), "\nFlags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Output version and bomb out
	if *flagVersion {
		config.PrintVersion(flag.CommandLine.Output())
		os.Exit(0)
	}

	// Check for less than one argument
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(0)
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

	// Publish messages
	topic := flag.Arg(0)
	if flag.NArg() > 1 {
		for _, data := range flag.Args()[1:] {
			if err := app.Publish(topic, data); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	} else {
		fmt.Printf("Reading messages from stdin, press CTRL+C to exit\n")
		bufio := bufio.NewReader(os.Stdin)
		for {
			line, err := bufio.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			if err := app.Publish(topic, line); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

	if err := app.Close(); err != nil {
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
