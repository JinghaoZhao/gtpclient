# gtpclient

This repo provide a GTP-U client for GTP-U module testing.

The gtpclient could generate UE traffic with assigned srcIP/dstIP/TEID, etc. Currently it supports the following modes:

- GTP: send & recv GTP-U packet (UDP traffic wrapped by GTP-U header), for mock gNB testing. E.g.,\
  `./gtpclient -cliip="127.0.0.1" -srvip="127.0.0.2" -srvport=2152 -uedstip="10.250.0.1"`
- UDP: send UDP packet with assigned UE IP, for SGI testing. E.g., \
  `./gtpclient -cliip="20.20.20.100" -srvip="20.20.20.2" -uedstip="10.250.0.1"`
- Echo: send echo request, for mock gNB testing. E.g., \
  `./gtpclient -mode="echo" -srvip="127.0.0.2" -cliport=8877`
- Recv: only recv incoming packets, for mock gNB testing. E.g., \
  `./gtpclient -mode="recv"`

### Build the gtpclient
To build the gtpclient, run `go build ./`

### Usage of gtpclient:
```
  -cliip string
        GTP Client IP (default "127.0.0.1")
  -cliport uint
        GTP Client port (default 2152)
  -mode string
        Test traffic type, support 'gtp', 'udp', 'echo', 'recv' (default "gtp")
  -srvip string
        GTP Server IP (default "127.0.0.1")
  -srvport uint
        GTP Server port (default 2152)
  -teid uint
        UE Session TEID (default 1234)
  -uedstip string
        UE Traffic Destination IP (default "10.0.0.2")
  -uedstport uint
        UE Traffic Destination Port (default 5678)
  -uesrcip string
        UE Traffic Source IP (default "10.0.0.1")
```
