package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"time"
)

func main() {
	// In the `jetstream` package, almost all API calls rely on `context.Context` for timeout/cancellation handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	nc, e1 := nats.Connect(nats.DefaultURL)
	fmt.Println("e1", e1)

	// Create a JetStream management interface
	js, e2 := jetstream.New(nc)
	fmt.Println("e2", e2)

	x, e3 := js.Stream(context.TODO(), "ORDERS2")
	fmt.Println("e3", e3, errors.Is(e3, jetstream.ErrStreamNotFound))
	printStreamState(ctx, x)

	// Create a stream
	s, e4 := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     "ORDERS2",
		Subjects: []string{"ORDERS2.*"},
	})
	fmt.Println("e4", e4)
	printStreamState(ctx, s)
}

func printStreamState(ctx context.Context, stream jetstream.Stream) {
	info, _ := stream.Info(ctx)
	if info != nil {
		b, _ := json.MarshalIndent(info.State, "", " ")
		fmt.Println(string(b))
	}
}
