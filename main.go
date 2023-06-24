package main

import (
	"context"
	"flag"
	multilog "github.com/umegbewe/kubectl-multilog/pkg"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {

	var kubeconfig *string

	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	contexts := flag.String("context", "", "context to use")
	namespace := flag.String("namespace", "default", "namespace to use")
	logLevel := flag.String("log-level", "info", "log level to use")
	//container := flag.String("container", "", "container to use")
	selector := flag.String("selector", "", "label selector to use")
	initContainers := flag.Bool("init-containers", false, "include init containers")
	previous := flag.Bool("previous", false, "include previous terminated containers")
	flag.Parse()

	pareseLogLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Fatalf("Invalid log level: %v", err)
	}

	logger := logrus.New()
	logger.SetLevel(pareseLogLevel)

	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		logger.Info("Received termination, stopping logs streams.....")
		cancel()
	}()

	kubeContext := strings.Split(*contexts, ",")

	var wg sync.WaitGroup

	for _, context := range kubeContext {
		wg.Add(1)
		go func(context string) {
			defer wg.Done()
			err = multilog.StreamLogs(ctx, logger, *kubeconfig, context, *namespace, *selector, *initContainers, *previous)
			if err != nil {
				logger.Errorf("Error streaming logs from context %s: %v", kubeContext, err)
			}
		}(context)
	}

	wg.Wait()

}
