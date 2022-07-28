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
	"strconv"
	"time"
)

var (
	gtpCliIP             = flag.String("cliip", "127.0.0.1", "GTP Client IP")
	gtpCliPort           = flag.Uint("cliport", 2152, "GTP Client port")
	gtpSrvIP             = flag.String("srvip", "127.0.0.1", "GTP Server IP")
	gtpSrvPort           = flag.Uint("srvport", 2152, "GTP Server port")
	testTEID             = flag.Uint("teid", 1234, "UE Session TEID")
	testUETrafficSrcIP   = flag.String("uesrcip", "10.0.0.1", "UE Traffic Source IP")
	testUETrafficDstIP   = flag.String("uedstip", "10.0.0.2", "UE Traffic Destination IP")
	testUETrafficDstPort = flag.Uint("uedstport", 5678, "UE Traffic Destination Port")
	testPayload          = make([]byte, 1400)
	testPktMode          = flag.String("mode", "gtp", "Test traffic type, support 'gtp', 'udp', 'echo', 'recv'")
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

func gtpServe() {
	testGtpCliConf := GTPConf{
		SrvAddr: *gtpCliIP,
		Port:    strconv.Itoa(int(*gtpCliPort)),
	}

	gtpSrvAddr := &net.UDPAddr{
		IP:   net.ParseIP(*gtpSrvIP),
		Port: int(*gtpSrvPort),
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

func udpServe() {
	ueAddr := &net.UDPAddr{
		Port: int(*testUETrafficDstPort),
		IP:   net.ParseIP(*testUETrafficDstIP),
	}

	conn, err := net.DialUDP("udp", nil, ueAddr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	for {
		color.Green("==> Send UDP packet to %+v \n", ueAddr)
		conn.Write([]byte("Write an UDP packet for DL testing"))
		time.Sleep(1 * time.Second)
	}
}

func recvModeServe() {
	testGtpCliConf := GTPConf{
		SrvAddr: *gtpCliIP,
		Port:    strconv.Itoa(int(*gtpCliPort)),
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

	color.Green("GTP Server is running on %+v \n", gtpCli.srvConn.LocalAddr())

	for {
		time.Sleep(1 * time.Second)
	}
}

func echoServe() {
	testGtpCliConf := GTPConf{
		SrvAddr: *gtpCliIP,
		Port:    strconv.Itoa(int(*gtpCliPort)),
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

	// Wait for the test client setup
	time.Sleep(1 * time.Second)

	peerAddr := &net.UDPAddr{
		IP:   net.ParseIP(*gtpSrvIP),
		Port: 2152,
	}

	for {
		// Send the Echo Request from client to the GTP server
		color.Green("==> Send Echo Request packet to %+v \n", peerAddr)
		err := gtpCli.SendEchoRequest(peerAddr)
		if err != nil {
			color.Red("echoServe Error: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
}

func main() {

	flag.Parse()

	switch *testPktMode {
	case "gtp":
		gtpServe()
	case "udp":
		udpServe()
	case "recv":
		recvModeServe()
	case "echo":
		echoServe()
	}
}
