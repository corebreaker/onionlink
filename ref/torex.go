package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cretz/bine/control"
	"github.com/cretz/bine/tor"
	"github.com/ipsn/go-libtor"
)

func run() error {
	// Wait at most a few minutes to publish the service
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start tor with default config (can set start conf's DebugWriter to os.Stdout for debug logs)
	fmt.Println("Starting tor and fetching title of https://check.torproject.org, please wait a few seconds...")
	t, err := tor.Start(ctx, &tor.StartConf{ProcessCreator: libtor.Creator, DebugWriter: os.Stderr, DataDir: "/home/frederic/data-dir"})
	if err != nil {
		panic(err)
	}

	defer t.Close()

	t.DeleteDataDirOnClose = false
	t.StopProcessOnClose = true
	t.Control.SetConf(
		&control.KeyVal{Key: "NewCircuitPeriod", Val: "1"},
		&control.KeyVal{Key: "CircuitIdleTimeout", Val: "300"},
		&control.KeyVal{Key: "MaxCircuitDirtiness", Val: "300"},
	)

	// Wait at most a minute to start network and get
	if err := t.EnableNetwork(nil, true); err != nil {
		log.Panicf("Failed to start tor: %v", err)
	}

	// Make connection
	dialer, err := t.Dialer(nil, &tor.DialConf{
		SkipEnableNetwork: true,
	})

	if err != nil {
		return err
	}

	log.Println("=====> Client")
	httpClient := &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}}
	// Get /
	resp, err := httpClient.Get("https://google.com")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Grab the <title>
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Headers:", resp.Header)
	fmt.Println("Body:", string(d))
	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
