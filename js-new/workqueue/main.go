package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// https://natsbyexample.com/examples/jetstream/workqueue-stream/go
func main() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	js, _ := jetstream.New(nc)

	cfg := jetstream.StreamConfig{
		Name:      "EVENTS",
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{"events.>"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, e2 := js.CreateStream(ctx, cfg)
	fmt.Println("created the stream", e2)

	js.Publish(ctx, "events.us.page_loaded", nil)
	js.Publish(ctx, "events.eu.mouse_clicked", nil)
	js.Publish(ctx, "events.us.input_focused", nil)
	fmt.Println("published 3 messages")

	fmt.Println("# Stream info without any consumers")
	printStreamState(ctx, stream)

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "processor-1",
	})
	fmt.Println(err)

	msgs, _ := cons.Fetch(3)
	for msg := range msgs.Messages() {
		msg.DoubleAck(ctx)
	}

	fmt.Println("\n# Stream info with one consumer")
	printStreamState(ctx, stream)

	_, err = stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "processor-2",
	})
	fmt.Println("\n# Create an overlapping consumer")
	fmt.Println(err)

	stream.DeleteConsumer(ctx, "processor-1")

	_, err = stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "processor-2",
	})
	fmt.Printf("created the new consumer? %v\n", err == nil)
	stream.DeleteConsumer(ctx, "processor-2")

	fmt.Println("\n# Create non-overlapping consumers")
	cons1, _ := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          "processor-us",
		FilterSubject: "events.us.>",
	})
	cons2, _ := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          "processor-eu",
		FilterSubject: "events.eu.>",
	})

	js.Publish(ctx, "events.eu.mouse_clicked", nil)
	js.Publish(ctx, "events.us.page_loaded", nil)
	js.Publish(ctx, "events.us.input_focused", nil)
	js.Publish(ctx, "events.eu.page_loaded", nil)
	fmt.Println("published 4 messages")

	msgs, _ = cons1.Fetch(2)
	for msg := range msgs.Messages() {
		fmt.Printf("us sub got: %s\n", msg.Subject())
		msg.Ack()
	}

	msgs, _ = cons2.Fetch(2)
	for msg := range msgs.Messages() {
		fmt.Printf("eu sub got: %s\n", msg.Subject())
		msg.Ack()
	}
}

func printStreamState(ctx context.Context, stream jetstream.Stream) {
	info, _ := stream.Info(ctx)
	if info != nil {
		b, _ := json.MarshalIndent(info.State, "", " ")
		fmt.Println(string(b))
	}
}
