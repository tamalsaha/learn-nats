#!/bin/bash

# account todd-test-a
nsc add account todd-test-a
nsc edit account todd-test-a --js-mem-storage=-1 --js-disk-storage=-1 --js-streams=-1 --js-consumer=-1
nsc add user todd-test-a -a todd-test-a

# account todd-test-b
nsc add account todd-test-b
nsc edit account todd-test-b --js-mem-storage=-1 --js-disk-storage=-1 --js-streams=-1 --js-consumer=-1
nsc add user todd-test-b -a todd-test-b

# account todd-test-c
nsc add account todd-test-c
nsc edit account todd-test-c --js-mem-storage=-1 --js-disk-storage=-1 --js-streams=-1 --js-consumer=-1
nsc add user todd-test-c -a todd-test-c


nsc push -u nats://localhost:4222 -A

nats context add test-a --server localhost:4222 --creds $NKEYS_PATH/creds/A/todd-test-a/todd-test-a.creds
nats context add test-b --server localhost:4222 --creds $NKEYS_PATH/creds/A/todd-test-b/todd-test-b.creds
nats context add test-c --server localhost:4222 --creds $NKEYS_PATH/creds/A/todd-test-c/todd-test-c.creds


echo ""
echo 'ACCTA="todd-test-a"'
echo ACCTAPUBKEY=$(nsc describe account todd-test-a --json | jq .sub)

echo ""
echo 'ACCTB="todd-test-b"'
echo ACCTBPUBKEY=$(nsc describe account todd-test-b --json | jq .sub)

echo ""
echo 'ACCTC="todd-test-c"'
echo ACCTCPUBKEY=$(nsc describe account todd-test-c --json | jq .sub)

