package shared

import (
	"fmt"
	"github.com/rs/xid"
)

// const NATS_URL = nats.DefaultURL

const NATS_URL = "nats://this-is-nats.appscode.ninja:4222"

func ProxyHandlerSubject(clusterUID string) string {
	return fmt.Sprintf("k8s.%s.proxy.handler", clusterUID)
}

func ProxyStatusSubject(clusterUID string) string {
	return fmt.Sprintf("k8s.%s.proxy.status", clusterUID)
}

func ProxyResponseSubjects() (hubSub, edgeSub string) {
	prefix := "k8s.proxy.resp"
	uid := xid.New().String()
	sub := fmt.Sprintf("%s.%s", prefix, uid)
	return sub, sub
}
