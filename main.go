package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
)

const testServerUrl = "http://localhost:3000"
const providerUrl = "http://localhost:7080/realms/realm01"

var sample = flag.String("sample", "sample1", "sample")

func main() {
	flag.Parse()

	// Create base context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %v", sig)
		cancel()
	}()

	log.Println(*sample)
	switch *sample {
	case "sample1":
		if err := sample1(ctx); err != nil {
			log.Fatal(err)
		}
	case "noHeadless":
		if err := noHeadless(ctx); err != nil {
			log.Fatal(err)
		}
	case "loginAndApi":
		if err := loginAndApi(ctx); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Unknown sample: %s", *sample)
	}
}

func sample1(ctx context.Context) error {
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		ctx,
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	// navigate to a page, wait for an element, click
	var example string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://pkg.go.dev/time`),
		// wait for footer element is visible (ie, page is loaded)
		chromedp.WaitVisible(`body > footer`),
		// find and click "Example" link
		chromedp.Click(`#example-After`, chromedp.NodeVisible),
		// retrieve the text of the textarea
		chromedp.Value(`#example-After textarea`, &example),
	)
	if err != nil {
		return err
	}
	log.Printf("Go's time.After example:\n%s", example)

	return nil
}

func noHeadless(ctx context.Context) error {
	ctx, cancel := chromedp.NewExecAllocator(
		ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-gpu", false),
			chromedp.Flag("hide-scrollbars", false),
		)...,
	)
	defer cancel()

	// create chrome instance
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// navigate to a page, wait for an element, click
	var example string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://github.com/login`),
		chromedp.ActionFunc(func(context.Context) error {
			log.Printf(">>>>>>>>>>>>>>>>>>>> Waiting for user to login")
			return nil
		}),
		chromedp.WaitVisible(`.logged-in`),
		chromedp.ActionFunc(func(context.Context) error {
			log.Printf(">>>>>>>>>>>>>>>>>>>> Logged In")
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Go's time.After example:\n%s", example)
	return nil
}
