# Cluster Connector Reverse Export

# Run nats server

```
> nsc env
+----------------------------------------------------------------------------------------------------------+
|                                             NSC Environment                                              |
+--------------------+-----+-------------------------------------------------------------------------------+
| Setting            | Set | Effective Value                                                               |
+--------------------+-----+-------------------------------------------------------------------------------+
| $NSC_CWD_ONLY      | No  | If set, default operator/account from cwd only                                |
| $NSC_NO_GIT_IGNORE | No  | If set, no .gitignore files written                                           |
| $NKEYS_PATH        | No  | ~/.local/share/nats/nsc/keys                                                  |
| $NSC_HOME          | No  | ~/.config/nats/nsc                                                            |
| $NATS_CA           | No  | If set, root CAs in the referenced file will be used for nats connections     |
|                    |     | If not set, will default to the system trust store                            |
| $NATS_KEY          | No  | If set, the tls key in the referenced file will be used for nats connections  |
| $NATS_CERT         | No  | If set, the tls cert in the referenced file will be used for nats connections |
+--------------------+-----+-------------------------------------------------------------------------------+
| From CWD           |     | No                                                                            |
| Default Stores Dir |     | ~/.local/share/nats/nsc/stores                                                |
| Current Store Dir  |     | ~/.local/share/nats/nsc/stores                                                |
| Current Operator   |     | appscode                                                                      |
| Current Account    |     | B                                                                             |
| Root CAs to trust  |     | Default: System Trust Store                                                   |
+--------------------+-----+-------------------------------------------------------------------------------+
```

```
> nsc init
? enter a configuration directory /Users/tamal/.local/share/nats/nsc/stores
? Select an operator Create Operator
? name your operator, account and user appscode
[ OK ] created operator appscode
[ OK ] created system_account: name:SYS id:AD3SU3IQ4TJFYCUPNEBPQXU7GZUJ76C6I7XEXEZVKMDKZMQUSQOU43GP
[ OK ] created system account user: name:sys id:UAIPNSRNPZ76IXJVWO3KDJ33JCEIQ6YVZSYVZW7ZBHQ6WGP25LHYOKAA
[ OK ] system account user creds file stored in `~/.local/share/nats/nsc/keys/creds/appscode/SYS/sys.creds`
[ OK ] created account appscode
[ OK ] created user "appscode"
[ OK ] project jwt files created in `~/.local/share/nats/nsc/stores`
[ OK ] user creds file stored in `~/.local/share/nats/nsc/keys/creds/appscode/appscode/appscode.creds`
> to run a local server using this configuration, enter:
>   nsc generate config --mem-resolver --config-file <path/server.conf>
> then start a nats-server using the generated config:
>   nats-server -c <path/server.conf>
all jobs succeeded
```

```
> nsc generate config --config-file server.conf --nats-resolver
[ OK ] wrote server configuration to `~/go/src/github.com/tamalsaha/learn-nats/server.conf`
Success!! - generated `~/go/src/github.com/tamalsaha/learn-nats/server.conf`
```

```
> nats-server -js -m 8222 -c server.conf
```

```
> nsc push -u nats://localhost:4222
[ OK ] push to nats-server "nats://localhost:4222" using system account "SYS":
       [ OK ] push appscode to nats-server with nats account resolver:
              [ OK ] pushed "appscode" to nats-server NCHCX25NIW2IXTKVFGUTPNNNTPIAPDUWQJBGGLBUWB7BRN4JWFTLMLRE: jwt updated
              [ OK ] pushed to a total of 1 nats-server
```

# A represents Admin account
```
> nsc add account A
[ OK ] generated and stored account key "ABBT43KI7PLDNI5555ZWDXJDSGWOO6UZ3SBIXXOI7EDILR67NUTSI4EL"
[ OK ] added account "A"

> nsc add user x -a A
[ OK ] generated and stored user key "UB3JM56ORWMFYMI6HP6W7RHMZNFRIRCSCDX3IR7NVHO6OZD6KDQC24WF"
[ OK ] generated user creds file `~/.local/share/nats/nsc/keys/creds/appscode/A/x.creds`
[ OK ] added user "x" to account "A"
```

# B is a Kubernetes cluster
```
> nsc add account B
[ OK ] generated and stored account key "AAT5B2HSGG3HACTVNA6FYXCVPSO6VRYHV44XTJZAPQY2PCARH2IUSVXD"
[ OK ] added account "B"
tamal@m1 ~/g/s/g/t/l/revsvc (master)> nsc add user y -a B
[ OK ] generated and stored user key "UB7E6UJI7Q2MQHLFGFYGHAWTL27KTB23VEFXWOZ3YMU36SAVW3C6GK4A"
[ OK ] generated user creds file `~/.local/share/nats/nsc/keys/creds/appscode/B/y.creds`
[ OK ] added user "y" to account "B"
```

```
> nsc push -u nats://localhost:4222 -A
[ OK ] push to nats-server "nats://localhost:4222" using system account "SYS":
       [ OK ] push A to nats-server with nats account resolver:
              [ OK ] pushed "A" to nats-server NCHCX25NIW2IXTKVFGUTPNNNTPIAPDUWQJBGGLBUWB7BRN4JWFTLMLRE: jwt updated
              [ OK ] pushed to a total of 1 nats-server
       [ OK ] push B to nats-server with nats account resolver:
              [ OK ] pushed "B" to nats-server NCHCX25NIW2IXTKVFGUTPNNNTPIAPDUWQJBGGLBUWB7BRN4JWFTLMLRE: jwt updated
              [ OK ] pushed to a total of 1 nats-server
       [ OK ] push SYS to nats-server with nats account resolver:
              [ OK ] pushed "SYS" to nats-server NCHCX25NIW2IXTKVFGUTPNNNTPIAPDUWQJBGGLBUWB7BRN4JWFTLMLRE: jwt updated
              [ OK ] pushed to a total of 1 nats-server
       [ OK ] push appscode to nats-server with nats account resolver:
              [ OK ] pushed "appscode" to nats-server NCHCX25NIW2IXTKVFGUTPNNNTPIAPDUWQJBGGLBUWB7BRN4JWFTLMLRE: jwt updated
              [ OK ] pushed to a total of 1 nats-server
```


# A exports stream k8s.proxy.handler.>

- https://docs.nats.io/using-nats/nats-tools/nsc/streams

## create private stream export
```
nsc add export --subject "k8s.proxy.handler.*" --private -a A

nsc describe account A
```

## create activation key for B

```
> nsc list keys --account B
+--------------------------------------------------------------------------------------------+
|                                            Keys                                            |
+----------+----------------------------------------------------------+-------------+--------+
| Entity   | Key                                                      | Signing Key | Stored |
+----------+----------------------------------------------------------+-------------+--------+
| appscode | OBWAE2E6YXFNJSGHT4WOL5SX77VSDQAAT6M5DD4KFFTHW2R34CWJUMXP |             | *      |
|  B       | AAT5B2HSGG3HACTVNA6FYXCVPSO6VRYHV44XTJZAPQY2PCARH2IUSVXD |             | *      |
|   y      | UB7E6UJI7Q2MQHLFGFYGHAWTL27KTB23VEFXWOZ3YMU36SAVW3C6GK4A |             | *      |
+----------+----------------------------------------------------------+-------------+--------+

> nsc generate activation \
  --account A \
  --target-account AAT5B2HSGG3HACTVNA6FYXCVPSO6VRYHV44XTJZAPQY2PCARH2IUSVXD \
  --subject k8s.proxy.handler.cid_b \
  -o /tmp/activation.jwt

> nsc add import --account B --token /tmp/activation.jwt --local-subject k8s.proxy.handler
[ OK ] added stream import "k8s.proxy.handler.cid_b"

> nsc push -u nats://localhost:4222 -A

> nsc describe account B
+--------------------------------------------------------------------------------------+
|                                   Account Details                                    |
+---------------------------+----------------------------------------------------------+
| Name                      | B                                                        |
| Account ID                | AAT5B2HSGG3HACTVNA6FYXCVPSO6VRYHV44XTJZAPQY2PCARH2IUSVXD |
| Issuer ID                 | OBWAE2E6YXFNJSGHT4WOL5SX77VSDQAAT6M5DD4KFFTHW2R34CWJUMXP |
| Issued                    | 2022-10-13 05:06:13 UTC                                  |
| Expires                   |                                                          |
+---------------------------+----------------------------------------------------------+
| Max Connections           | Unlimited                                                |
| Max Leaf Node Connections | Unlimited                                                |
| Max Data                  | Unlimited                                                |
| Max Exports               | Unlimited                                                |
| Max Imports               | Unlimited                                                |
| Max Msg Payload           | Unlimited                                                |
| Max Subscriptions         | Unlimited                                                |
| Exports Allows Wildcards  | True                                                     |
| Response Permissions      | Not Set                                                  |
+---------------------------+----------------------------------------------------------+
| Jetstream                 | Disabled                                                 |
+---------------------------+----------------------------------------------------------+
| Exports                   | None                                                     |
+---------------------------+----------------------------------------------------------+

+------------------------------------------------------------------------------------------------------------------+
|                                                     Imports                                                      |
+-------------------------+--------+-------------------------+-------------------+---------+--------------+--------+
| Name                    | Type   | Remote                  | Local             | Expires | From Account | Public |
+-------------------------+--------+-------------------------+-------------------+---------+--------------+--------+
| k8s.proxy.handler.cid_b | Stream | k8s.proxy.handler.cid_b | k8s.proxy.handler |         | A            | No     |
+-------------------------+--------+-------------------------+-------------------+---------+--------------+--------+
```

## Test stream export/import

```
> nsc push -u nats://localhost:4222 -A

# from nsc env
NKEYS_PATH=$HOME/.local/share/nats/nsc/keys

# B
> nats sub --creds=$NKEYS_PATH/creds/appscode/B/y.creds k8s.proxy.handler

# A
> nats pub --creds=$NKEYS_PATH/creds/appscode/A/x.creds k8s.proxy.handler.cid_b hello
```


# A exports service k8s.proxy.resp.>

- https://docs.nats.io/using-nats/nats-tools/nsc/services#creating-a-private-service-export

## Creating a Private Service Export

```
nsc add export --subject "k8s.proxy.resp.>" --private --service --account A
```

## Generating an Activation Token

```
nsc generate activation \
  --account A \
  --target-account AAT5B2HSGG3HACTVNA6FYXCVPSO6VRYHV44XTJZAPQY2PCARH2IUSVXD \
  --subject k8s.proxy.resp.cid_b.* \
  -o /tmp/activation.jwt
```

## Importing a Private Service

```
> nsc add import --account B -u /tmp/activation.jwt \
  --local-subject k8s.proxy.resp.* \
  --name k8s.proxy.resp

> nsc push -u nats://localhost:4222 -A

> nsc describe account B
```

## Test service export/import

```
> nsc push -u nats://localhost:4222 -A

# from nsc env
NKEYS_PATH=$HOME/.local/share/nats/nsc/keys

# B
> nats sub --creds=$NKEYS_PATH/creds/appscode/B/y.creds k8s.proxy.handler

# A
> nats pub --creds=$NKEYS_PATH/creds/appscode/A/x.creds k8s.proxy.handler.cid_b hello
```

```
# A
> nats reply --creds=$NKEYS_PATH/creds/appscode/A/x.creds k8s.proxy.resp.cid_b.* "help is here"

# B
> nats req --creds=$NKEYS_PATH/creds/appscode/B/y.creds k8s.proxy.resp.1 help_me
```

## Combined mode

- https://github.com/nats-io/nats.docs/blob/master/using-nats/developing-with-nats/sending/replyto.md

```
> go run reply/main.go
REQ: hello REPLY_TO: k8s.proxy.resp.cd3sfv27qo0lcesii46g

```

```
> go run req/main.go
2022/10/13 01:10:36 Reply: echo>>>hello
```
