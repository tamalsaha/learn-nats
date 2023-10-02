package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

// https://natsbyexample.com/examples/jetstream/workqueue-stream/go
// wget https://github.com/tamalsaha/learn-nats/raw/master/js-new/wait_for_gha_run/wait_for_gha_run
// chmod +x wait_for_gha_run
func main() {
	var url = flag.String("nats-addr", nats.DefaultURL, "NATS serve address")
	flag.Parse()

	nc, err := NewConnection(*url, "")
	if err != nil {
		klog.Fatalln(err)
	}
	defer nc.Drain()

	for {
		err := wait_until_job(nc)
		if err != nil {
			klog.ErrorS(err, "error while waiting for next job")
		}
		time.Sleep(10 * time.Second)
	}
}

func wait_until_job(nc *nats.Conn) error {
	js, err := jetstream.New(nc)
	if err != nil {
		return err
	}

	ctx := context.TODO()

	streamName := "gha_queued"
	streamQueued, err := js.Stream(ctx, streamName)
	if err != nil {
		return err
	}
	klog.Info("found the stream", streamName)

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	defer printStreamState(ctx, streamQueued)

	err = consumeMsg(ctx, streamQueued, hostname, streamName+".high")
	if err == nil {
		return nil
	}
	return consumeMsg(ctx, streamQueued, hostname, streamName+".regular")
}

func consumeMsg(ctx context.Context, streamQueued jetstream.Stream, consumer, subj string) error {
	cons, err := streamQueued.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          consumer,
		FilterSubject: subj,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}
	defer streamQueued.DeleteConsumer(ctx, consumer)

	/*
			Double-acking is a mechanism used in JetStream to ensure exactly once semantics in message processing.
		    It involves calling the `AckSync()` function instead of `Ack()` to set a reply subject on the Ack and
		    wait for a response from the server on the reception and processing of the acknowledgement. This helps to
		    avoid message duplication and guarantees that the message will not be re-delivered by the consumer.
	*/
	msgs, err := cons.FetchNoWait(1)
	if err != nil {
		return err
	}
	for msg := range msgs.Messages() {
		if err := msg.DoubleAck(ctx); err != nil {
			return err
		} else {
			return nil // DONE
		}
	}
	if msgs.Error() != nil {
		return errors.Wrap(msgs.Error(), "error during Fetch()")
	}
	return nil
}

func printStreamState(ctx context.Context, stream jetstream.Stream) {
	info, _ := stream.Info(ctx)
	if info != nil {
		b, _ := json.MarshalIndent(info.State, "", " ")
		fmt.Println(string(b))
	}
}
