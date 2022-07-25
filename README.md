# gtpclient

This repo provide a GTP-U client for GTP-U module testing.

The gtpclient could generate UE traffic with assigned srcIP/dstIP/TEID, etc. 

To build the gtpclient, run `go build ./`

Usage of ./gtpclient:\
  -cliip string GTP Client IP (default "127.0.0.1")\
  -cliport string GTP Client port (default "2152")\
  -srvip string GTP Server IP (default "127.0.0.1")\
  -teid uint UE Session TEID (default 1234)\
  -uedstip string UE Traffic Destination IP (default "10.0.0.2")\
  -uesrcip string UE Traffic Source IP (default "10.0.0.1")
