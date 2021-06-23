# nats-hop-demo

```
$ nats-server

$ nats sub cli.demo -s nats://nats-hop:4222

$ nats pub cli.demo "message {{.Count}} @ {{.TimeStamp}}" --count=10 -s nats://nats-hop:4222
```
