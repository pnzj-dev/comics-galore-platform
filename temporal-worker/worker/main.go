// Command worker runs the Temporal worker that hosts the Comics Galore
// checkout/subscription/boost workflows and their activities.
package main

import (
	"crypto/tls"
	"log"
	"os"

	"comics-galore/temporal-worker/activities"
	"comics-galore/temporal-worker/workflows"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	address := envOr("TEMPORAL_ADDRESS", "localhost:7233")
	namespace := envOr("TEMPORAL_NAMESPACE", "default")
	backendURL := envOr("ENCORE_BACKEND_URL", "http://localhost:4000")
	workerSecret := os.Getenv("WORKER_SECRET")

	opts := client.Options{HostPort: address, Namespace: namespace}
	switch {
	case os.Getenv("TEMPORAL_API_KEY") != "":
		opts.Credentials = client.NewAPIKeyStaticCredentials(os.Getenv("TEMPORAL_API_KEY"))
	default:
		if certPEM, keyPEM := os.Getenv("TEMPORAL_CERT"), os.Getenv("TEMPORAL_KEY"); certPEM != "" && keyPEM != "" {
			cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
			if err != nil {
				log.Fatalf("temporal tls: %v", err)
			}
			opts.ConnectionOptions = client.ConnectionOptions{
				TLS: &tls.Config{Certificates: []tls.Certificate{cert}},
			}
		}
	}

	c, err := client.Dial(opts)
	if err != nil {
		log.Fatalf("temporal dial: %v", err)
	}
	defer c.Close()

	w := worker.New(c, workflows.TaskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.SubscribeWorkflow)
	w.RegisterWorkflow(workflows.BoostWorkflow)
	w.RegisterActivity(&activities.Activities{BackendURL: backendURL, WorkerSecret: workerSecret})

	log.Printf("temporal worker started: address=%s namespace=%s backend=%s", address, namespace, backendURL)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
