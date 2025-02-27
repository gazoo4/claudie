package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Berops/claudie/internal/envs"
	"github.com/Berops/claudie/internal/healthcheck"
	"github.com/Berops/claudie/internal/utils"
	"github.com/Berops/claudie/internal/worker"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/Berops/claudie/proto/pb"
)

const defaultBuilderPort = 50051

// healthCheck function is function used for querying readiness of the pod running this microservice
func healthCheck() error {
	//Check if Builder can connect to Terraformer/Ansibler/Kube-eleven/Kuber
	//Connection to these services are crucial for Builder, without them, the builder is NOT Ready
	if cc, err := utils.GrpcDialWithInsecure("terraformer", envs.TerraformerURL); err != nil {
		return err
	} else {
		if err := cc.Close(); err != nil {
			log.Error().Msgf("Error closing the connection in health check function : %v", err)
		}
	}
	if cc, err := utils.GrpcDialWithInsecure("ansibler", envs.AnsiblerURL); err != nil {
		return err
	} else {
		if err := cc.Close(); err != nil {
			log.Error().Msgf("Error closing the connection in health check function : %v", err)
		}
	}
	if cc, err := utils.GrpcDialWithInsecure("kubeEleven", envs.KubeElevenURL); err != nil {
		return err
	} else {
		if err := cc.Close(); err != nil {
			log.Error().Msgf("Error closing the connection in health check function : %v", err)
		}
	}
	if cc, err := utils.GrpcDialWithInsecure("kuber", envs.KuberURL); err != nil {
		return err
	} else {
		if err := cc.Close(); err != nil {
			log.Error().Msgf("Error closing the connection in health check function : %v", err)
		}
	}
	return nil
}

func main() {
	// initialize logger
	utils.InitLog("builder")

	// Create connection to Context-box
	cc, err := utils.GrpcDialWithInsecure("context-box", envs.ContextBoxURL)
	log.Info().Msgf("Dial Context-box: %s", envs.ContextBoxURL)
	if err != nil {
		log.Fatal().Msgf("Could not connect to Content-box: %v", err)
	}
	defer func() { utils.CloseClientConnection(cc) }()
	// Creating the client
	c := pb.NewContextBoxServiceClient(cc)

	// Initialize health probes
	healthChecker := healthcheck.NewClientHealthChecker(fmt.Sprint(defaultBuilderPort), healthCheck)
	healthChecker.StartProbes()

	g, ctx := errgroup.WithContext(context.Background())
	w := worker.NewWorker(ctx, 5*time.Second, configProcessor(c), worker.ErrorLogger)
	// interrupt catching goroutine
	g.Go(func() error {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		defer signal.Stop(ch)
		<-ch
		return errors.New("builder interrupt signal")
	})
	//builder goroutine
	g.Go(func() error {
		w.Run()
		return nil
	})
	log.Info().Msgf("Stopping Builder: %v", g.Wait())
}
