# nsc-demo

```
> nsc init
? enter a configuration directory /Users/tamal/.local/share/nats/nsc/stores
? Select an operator Create Operator
? name your operator, account and user appscode
[ OK ] created operator appscode
[ OK ] created system_account: name:SYS id:AARMV3BBUGRIDJSF5T6HXPPBYCQ2WT53V7ACWKHMJCPWBR75MDNF7KOS
[ OK ] created system account user: name:sys id:UBWMI5XGYJTACK5QSEZE2GX5NFBQPCOT26WGNC3R6YNNEH5F6DG4ATCT
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


> nsc generate config --mem-resolver --config-file server.conf
[ OK ] wrote server configuration to `~/go/src/github.com/tamalsaha/learn-nats/nsc-demo/server.conf`
Success!! - generated `~/go/src/github.com/tamalsaha/learn-nats/nsc-demo/server.conf`


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
| Current Account    |     | appscode                                                                      |
| Root CAs to trust  |     | Default: System Trust Store                                                   |
+--------------------+-----+-------------------------------------------------------------------------------+

> ls -l ~/.local/share/nats/nsc
> ls -l ~/.config/nats/nsc


> nsc list keys -A
> nsc list keys -A --show-seeds
```

```
> nsc describe operator

> nsc describe account SYS
> nsc describe user sys -a SYS

> nsc describe account appscode
> nsc describe user appscode -a appscode
```

