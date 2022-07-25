package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/wmnsk/go-gtp/gtpv1"
	"github.com/wmnsk/go-gtp/gtpv1/message"
	"net"
	"time"
)

var (
	gtpCliIP           = flag.String("cliip", "127.0.0.1", "GTP Client IP")
	gtpCliPort         = flag.String("cliport", "2152", "GTP Client port")
	gtpSrvIP           = flag.String("srvip", "127.0.0.1", "GTP Server IP")
	testTEID           = flag.Uint("teid", 1234, "UE Session TEID")
	testUETrafficSrcIP = flag.String("uesrcip", "10.0.0.1", "UE Traffic Source IP")
	testUETrafficDstIP = flag.String("uedstip", "10.0.0.2", "UE Traffic Destination IP")
	testPayload        = make([]byte, 1400)
)

func newTestIPPacket(srcIP, dstIP string, payload []byte) (packet []byte, err error) {
	// Build the packet headers
	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	buffer := gopacket.NewSerializeBuffer()
	ipLayer := &layers.IPv4{
		Version:  4,
		TTL:      64,
		SrcIP:    net.ParseIP(srcIP),
		DstIP:    net.ParseIP(dstIP),
		Protocol: layers.IPProtocolUDP,
	}
	udpLayer := &layers.UDP{
		SrcPort: layers.UDPPort(5678),
		DstPort: layers.UDPPort(5678),
	}

	err = udpLayer.SetNetworkLayerForChecksum(ipLayer)
	if err != nil {
		return nil, err
	}

	// And create the packet with the layers
	err = gopacket.SerializeLayers(buffer, options,
		ipLayer,
		udpLayer,
		gopacket.Payload(payload),
	)
	if err != nil {
		return nil, err
	}

	packet = buffer.Bytes()
	return
}

func (gtpCli *GTPServer) AddTPDUHandler() {
	gtpCli.srvConn.AddHandler(message.MsgTypeTPDU, func(c gtpv1.Conn, senderAddr net.Addr, msg message.Message) error {
		var ip4 layers.IPv4
		pdu, ok := msg.(*message.TPDU)
		if !ok {
			fmt.Println("got unexpected type of message, should be TPDU")
			return errors.New("got unexpected type of message, should be TPDU")
		}

		parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ip4)
		decoded := []gopacket.LayerType{}
		parser.DecodeLayers(pdu.Payload, &decoded)

		for _, layerType := range decoded {
			switch layerType {
			case layers.LayerTypeIPv4:
				color.Blue("<== Successfully receive paging packet from %+v, msg={TEID=%d, length=%d, UE_IP=%s}\n", senderAddr, msg.TEID(), msg.MarshalLen(), ip4.DstIP)
			default:
				fmt.Println("The incoming packet header is not supported yet")
			}
		}
		return nil
	})
}

func main() {

	flag.Parse()

	testGtpCliConf := GTPConf{
		SrvAddr: *gtpCliIP,
		Port:    *gtpCliPort,
	}

	gtpSrvAddr := &net.UDPAddr{
		IP:   net.ParseIP(*gtpSrvIP),
		Port: 2152,
		Zone: "",
	}

	gtpCli, err := NewGTPServer(testGtpCliConf)
	if err != nil {
		fmt.Printf("NewGTPServer Error: %v", err)
	}
	gtpCli.AddTPDUHandler()

	go func() {
		err := gtpCli.Serve()
		if err != nil {
			panic(err)
		}
	}()
	defer gtpCli.Stop()

	// Wait for the test client and server setup
	time.Sleep(1 * time.Second)

	outgoingPacket, err := newTestIPPacket(*testUETrafficSrcIP, *testUETrafficDstIP, testPayload)
	if err != nil {
		panic(err)
	}

	for {
		color.Green("==> Send packet to %+v \n", gtpSrvAddr)
		if _, err = gtpCli.srvConn.WriteToGTP(uint32(*testTEID), outgoingPacket, gtpSrvAddr); err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
	}

}
