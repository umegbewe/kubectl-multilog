package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	multilog "github.com/umegbewe/kubectl-multilog/pkg"

	"github.com/sirupsen/logrus"
)

func main() {

	var kubeconfig *string

	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	contexts := flag.String("context", "", "context to use, comma seperated")
	namespace := flag.String("namespace", "default", "namespace to use, comma seperated")
	logLevel := flag.String("log-level", "info", "log level to use")
	container := flag.String("container", "", "container to use, comma seperated")
	selector := flag.String("selector", "", "label selector to use, comma seperated")
	tailLines := flag.Int64("n", 100, "number of lines to tail")
	previous := flag.Bool("previous", false, "include previous terminated containers")
	flag.Parse()

	pareseLogLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Fatalf("Invalid log level: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(pareseLogLevel)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		logger.Info("Received termination, stopping logs streams.....")
		cancel()

		// Give the streams some time to stop gracefully
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	kubeContext := strings.Split(*contexts, ",")
	selectors := strings.Split(*selector, ",")
	containers := strings.Split(*container, ",")
	namespaces := strings.Split(*namespace, ",")

	var wg sync.WaitGroup

	for _, context := range kubeContext {
		wg.Add(1)
		go func(context string) {
			defer wg.Done()
			err = multilog.StreamLogs(ctx, logger, *kubeconfig, context, namespaces, selectors, containers, *previous, *tailLines)
			if err != nil {
				logger.Errorf("Error streaming logs from context %s: %v", context, err)
			}
		}(context)
	}

	wg.Wait()

}
