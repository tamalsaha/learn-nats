# Subject mapping

```
nats-server -m 8222 -js
```

```
nats server mapping "ghactions.runs.*.*" "ghactions.machines.{{wildcard(1)}}.{{partition(3,2)}}"
```

```
> nats str create
? Stream Name ghactions
? Subjects ghactions.machines.*.*
? Storage file
? Replication 1
? Retention Policy Work Queue
? Discard Policy Old
? Stream Messages Limit -1
? Per Subject Messages Limit -1
? Total Stream Size -1
? Message TTL -1
? Max Message Size -1
? Duplicate tracking time window 2m0s
? Allow message Roll-ups No
? Allow message deletion Yes
? Allow purging subjects or the entire stream Yes
Stream ghactions was created
```

```
> nats request ghactions.runs.queued.1 "I need help!"

07:45:26 Sending request on "ghactions.runs.queued.1"
07:45:26 No responders are available
```
