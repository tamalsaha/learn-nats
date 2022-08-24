#!/bin/bash

# Account Refs
ACCTA="todd-test-a"
ACCTAPUBKEY="AAF2XH27TJA5H2LG5FMETB5IHKSBRYB6OL726IAOBOD3GDDWAL5XSGID"

ACCTB="todd-test-b"
ACCTBPUBKEY="ADWEEXK7UCQBS7FHITOBPIRY4B5CXZZBDN2MZM5NISXVJHR2YEWN2JZC"

ACCTC="todd-test-c"
ACCTCPUBKEY="ADK6KAS5GP75XTCNMBCEJJ3ARUQ7PEAQE2CTM7FZK6HBGGURPRQ7XMZU"

#### Exports ####

addexports () {
# ACCTA
nsc add export --private --account $ACCTA --name "ORDEREVENTS-GRANT-DELIVER" --subject "deliver.retail.v1.order.events"
nsc add export --private --account $ACCTA --name "ORDEREVENTS-GRANT-ACK" --subject "\$JS.ACK.ORDEREVENTS.ORDEREVENTS-C1.>" --service
nsc add export --private --account $ACCTA --name "ORDEREVENTS-GRANT-INFO" --subject "\$JS.API.CONSUMER.INFO.ORDEREVENTS.ORDEREVENTS-C1" --service

# Generate an activation token for ACCTB import
nsc generate activation --output-file "ORDEREVENTS-GRANT-DELIVER-ACCTB.tok" --account $ACCTA --subject "deliver.retail.v1.order.events" --target-account $ACCTBPUBKEY
nsc generate activation --output-file "ORDEREVENTS-GRANT-ACK-ACCTB.tok" --account $ACCTA --subject "\$JS.ACK.ORDEREVENTS.ORDEREVENTS-C1.>" --target-account $ACCTBPUBKEY
nsc generate activation --output-file "ORDEREVENTS-GRANT-INFO-ACCTB.tok" --account $ACCTA --subject "\$JS.API.CONSUMER.INFO.ORDEREVENTS.ORDEREVENTS-C1" --target-account $ACCTBPUBKEY
}

deleteexports () {
nsc delete export --account $ACCTA --subject "deliver.retail.v1.order.events"
nsc delete export --account $ACCTA --subject "\$JS.ACK.ORDEREVENTS.ORDEREVENTS-C1.>"
nsc delete export --account $ACCTA --subject "\$JS.API.CONSUMER.INFO.ORDEREVENTS.ORDEREVENTS-C1"
}

while getopts "a d" option
do
    case ${option} in
        d) deleteexports ;;
        a) addexports ;;
    esac
done