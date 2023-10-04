package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/olekukonko/tablewriter"
	"github.com/tamalsaha/learn-nats/natsclient"
	"github.com/tamalsaha/learn-nats/shared"
	"k8s.io/apimachinery/pkg/util/duration"
	"sort"
	"strconv"
	"time"
)

func main() {
	ConvertToHumanReadableDateType(nil)

	nc, err := natsclient.NewConnection(shared.NATS_URL, "")
	if err != nil {
		panic(err)
	}
	defer nc.Drain()

	names := []string{
		"gha_queued",
		"gha_completed",
	}
	streams, err := CollectStreamInfo(nc, names)
	if err != nil {
		panic(err)
	}
	data := RenderStreamInfo(streams)
	fmt.Println(string(data))

	for _, name := range names {
		fmt.Printf("\n\n## Consumers for Stream: %s\n", name)
		consumers, err := CollectConsumerInfo(nc, name)
		if err != nil {
			panic(err)
		}
		data := RenderConsumerInfo(consumers)
		fmt.Println(string(data))
	}
}

func CollectStreamInfo(nc *nats.Conn, names []string) ([]*jetstream.StreamInfo, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	result := make([]*jetstream.StreamInfo, 0, len(names))
	for _, name := range names {
		if s, err := js.Stream(ctx, name); err != nil {
			return nil, err
		} else {
			info, err := s.Info(ctx)
			if err != nil {
				return nil, err
			}
			if info != nil {
				result = append(result, info)
			}
		}
	}
	return result, nil
}

func RenderStreamInfo(streams []*jetstream.StreamInfo) []byte {
	data := make([][]string, 0, len(streams))
	for _, s := range streams {
		data = append(data, []string{
			s.Config.Name,
			ConvertToHumanReadableDateType(&s.Created),
			strconv.FormatUint(s.State.Msgs, 10),
			ConvertToHumanReadableDateType(&s.State.LastTime),
		})
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i][0] < data[j][0]
	})

	var buf bytes.Buffer

	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Name", "Created", "Messages", "Last Message"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return buf.Bytes()
}

// ConvertToHumanReadableDateType returns the elapsed time since timestamp in
// human-readable approximation.
// ref: https://github.com/kubernetes/apimachinery/blob/v0.21.1/pkg/api/meta/table/table.go#L63-L70
// But works for timestamp before or after now.
func ConvertToHumanReadableDateType(timestamp *time.Time) string {
	if timestamp == nil || timestamp.IsZero() {
		return "<unknown>"
	}
	var d time.Duration
	now := time.Now()
	if now.After(*timestamp) {
		d = now.Sub(*timestamp)
	} else {
		d = timestamp.Sub(now)
	}
	return duration.HumanDuration(d)
}

func CollectConsumerInfo(nc *nats.Conn, streamName string) ([]*jetstream.ConsumerInfo, error) {
	ctx := context.TODO()
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	s, err := js.Stream(ctx, streamName)
	if err != nil {
		return nil, err
	}

	var result []*jetstream.ConsumerInfo

	consumers := s.ListConsumers(ctx)
	for cons := range consumers.Info() {
		result = append(result, cons)
	}
	if consumers.Err() != nil {
		return nil, err
	}
	return result, nil
}

func RenderConsumerInfo(consumers []*jetstream.ConsumerInfo) []byte {
	data := make([][]string, 0, len(consumers))
	for _, s := range consumers {

		data = append(data, []string{
			s.Config.Name,
			ConvertToHumanReadableDateType(&s.Created),
			strconv.FormatBool(!s.PushBound),
			s.Config.FilterSubject,
			ConvertToHumanReadableDateType(s.Delivered.Last),
			ConvertToHumanReadableDateType(s.AckFloor.Last),
			strconv.Itoa(s.NumAckPending),
			strconv.Itoa(s.NumRedelivered),
			strconv.Itoa(s.NumWaiting),
			strconv.FormatUint(s.NumPending, 10),
		})
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i][0] < data[j][0]
	})

	var buf bytes.Buffer

	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Name", "Created", "Pull", "Filter Subject", "Last Delivery", "Last Ack", "NumAckPending", "NumRedelivered", "NumWaiting", "NumPending"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	return buf.Bytes()
}
