package shared

import "fmt"

// const NATS_URL = nats.DefaultURL

const NATS_URL = "nats://45.79.14.143:4222"

func ProxyHandlerSubject(clusterUID string) string {
	return fmt.Sprintf("k8s.%s.proxy.handler", clusterUID)
}

func ProxyStatusSubject(clusterUID string) string {
	return fmt.Sprintf("k8s.%s.proxy.status", clusterUID)
}
