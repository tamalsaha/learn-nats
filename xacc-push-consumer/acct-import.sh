#!/bin/bash

# Account Refs
ACCTA="todd-test-a"
ACCTAPUBKEY="AAF2XH27TJA5H2LG5FMETB5IHKSBRYB6OL726IAOBOD3GDDWAL5XSGID"

ACCTB="todd-test-b"
ACCTBPUBKEY="ADWEEXK7UCQBS7FHITOBPIRY4B5CXZZBDN2MZM5NISXVJHR2YEWN2JZC"

ACCTC="todd-test-c"
ACCTCPUBKEY="ADK6KAS5GP75XTCNMBCEJJ3ARUQ7PEAQE2CTM7FZK6HBGGURPRQ7XMZU"

#### Imports ####

addimports () {
# ACCTB
nsc add import --token "ORDEREVENTS-GRANT-DELIVER-ACCTB.tok" --account $ACCTB --name "ORDEREVENTS-GRANT-DELIVER" --local-subject "retail.v1.order.events"
nsc add import --token "ORDEREVENTS-GRANT-ACK-ACCTB.tok" --account $ACCTB --name "ORDEREVENTS-GRANT-ACK" --local-subject "\$JS.ACK.ORDEREVENTS.ORDEREVENTS-C1.>" --service
nsc add import --token "ORDEREVENTS-GRANT-INFO-ACCTB.tok" --account $ACCTB --name "ORDEREVENTS-GRANT-INFO" --local-subject "ACCTA.API.CONSUMER.INFO.ORDEREVENTS.ORDEREVENTS-C1" --service
}

deleteimports () {
nsc delete import --account $ACCTB --src-account $ACCTAPUBKEY --subject "deliver.retail.v1.order.events"
nsc delete import --account $ACCTB --src-account $ACCTAPUBKEY --subject "\$JS.ACK.ORDEREVENTS.ORDEREVENTS-C1.>"
nsc delete import --account $ACCTB --src-account $ACCTAPUBKEY --subject "\$JS.API.CONSUMER.INFO.ORDEREVENTS.ORDEREVENTS-C1"
}

while getopts "a d" option
do
    case ${option} in
        d) deleteimports ;;
        a) addimports ;;
    esac
done