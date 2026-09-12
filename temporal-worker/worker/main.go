// Command worker runs the Temporal worker that hosts the Comics Galore
// subscription workflow and its activities.
package main

import (
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

	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("temporal dial: %v", err)
	}
	defer c.Close()

	w := worker.New(c, workflows.TaskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.SubscriptionWorkflow)
	w.RegisterActivity(&activities.Activities{BackendURL: backendURL})

	log.Printf("temporal worker started: address=%s namespace=%s backend=%s", address, namespace, backendURL)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
