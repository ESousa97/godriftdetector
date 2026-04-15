// Package infra provides the infrastructure adapters for godriftdetector.
//
// It includes concrete implementations for reading the desired state from
// docker-compose files ([ComposeReader]), fetching the actual state from
// the Docker daemon ([DockerProvider]), syncing configurations from remote
// Git repositories ([GitProvider]), and sending alerts via webhooks ([WebhookNotifier]).
package infra
