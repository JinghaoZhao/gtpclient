package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/wmnsk/go-gtp/gtpv1"
	"github.com/wmnsk/go-gtp/gtpv1/ie"
	"github.com/wmnsk/go-gtp/gtpv1/message"
	"net"
	"sync"
)

// GTPConf provides basic configurations for GTP Server
type GTPConf struct {
	SrvAddr string `json:"srv_addr"`
	Port    string `json:"port"`
}

// GTPServer defines basic GTP connection and message handlers
type GTPServer struct {
	srvConn *gtpv1.UPlaneConn
	conf    GTPConf
	bufmap  sync.Map
}

// NewGTPServer create and config a GTP-U server and returns the server instance
func NewGTPServer(conf GTPConf) (gs *GTPServer, err error) {
	gtpSrvAddr, err := net.ResolveUDPAddr("udp", conf.SrvAddr+":"+conf.Port)
	if err != nil {
		fmt.Printf("Error while creating NewGTPServer: %v\n", err)
		return nil, err
	}

	srvConn := gtpv1.NewUPlaneConn(gtpSrvAddr)
	srvConn.DisableErrorIndication()

	gs = &GTPServer{
		srvConn: srvConn,
	}

	gs.AddEchoRequestHandler()
	gs.AddEchoResponseHandler()
	gs.AddEndMarkerHandler()
	//gs.AddPagingHandler()

	return gs, nil
}

// AddEchoRequestHandler processes Echo Request and responses Echo Response
func (gs *GTPServer) AddEchoRequestHandler() {
	gs.srvConn.AddHandler(message.MsgTypeEchoRequest, func(c gtpv1.Conn, senderAddr net.Addr, msg message.Message) error {
		if _, ok := msg.(*message.EchoRequest); !ok {
			fmt.Println("got unexpected type of message, should be Echo Request")
			return errors.New("got unexpected type of message, should be Echo Request")
		}

		fmt.Printf("Receive echo request packet from %+v, message=%+v\n", senderAddr, msg)
		// respond with EchoResponse.
		return c.RespondTo(
			senderAddr, msg, message.NewEchoResponse(0, ie.NewRecovery(c.Restarts())),
		)
	})
}

// AddEchoResponseHandler processes Echo Response
func (gs *GTPServer) AddEchoResponseHandler() {
	gs.srvConn.AddHandler(message.MsgTypeEchoResponse, func(c gtpv1.Conn, senderAddr net.Addr, msg message.Message) error {
		if _, ok := msg.(*message.EchoResponse); !ok {
			fmt.Println("got unexpected type of message, should be Echo Response")
			return errors.New("got unexpected type of message, should be Echo Response")
		}

		fmt.Printf("Receive echo response packet from %+v, message=%+v\n", senderAddr, msg)

		// do nothing now, leave for future peer monitoring
		return nil
	})
}

// NewEndMarker creates a new End Marker GTP packet.
func NewEndMarker(teid uint32, ies ...*ie.IE) *message.EndMarker {
	e := &message.EndMarker{
		// In 3GPP 29.281, End marker shall set the S flag to '0', thus the flag is 0x30
		Header: message.NewHeader(0x30, message.MsgTypeEndMarker, teid, 0, nil),
	}

	for _, i := range ies {
		if i == nil {
			continue
		}
		switch i.Type {
		case ie.PrivateExtension:
			e.PrivateExtension = i
		default:
			e.AdditionalIEs = append(e.AdditionalIEs, i)
		}
	}

	e.SetLength()
	return e
}

// SendEndMarker sends an End Marker to the peer address with corresponding TEID
func (gs *GTPServer) SendEndMarker(teid uint32, peerAddr net.Addr) error {
	b, err := NewEndMarker(teid).Marshal()
	if err != nil {
		fmt.Printf("Error while getting NewEndMarker: %v", err)
		return err
	}
	if _, err := gs.srvConn.WriteTo(b, peerAddr); err != nil {
		fmt.Printf("Error while writing End Marker to the peer address: %v", err)
		return err
	}
	return nil
}

// AddEndMarkerHandler processes incoming End Marker
func (gs *GTPServer) AddEndMarkerHandler() {
	gs.srvConn.AddHandler(message.MsgTypeEndMarker, func(c gtpv1.Conn, senderAddr net.Addr, msg message.Message) error {
		if _, ok := msg.(*message.EndMarker); !ok {
			fmt.Println("got unexpected type of message, should be End Marker")
			return errors.New("got unexpected type of message, should be End Marker")
		}

		fmt.Printf("Receive end marker packet from %+v, message=%+v\n", senderAddr, msg)
		// do nothing now, leave for future testing
		return nil
	})
}

// Serve starts GTP-U server instance
// blocking call
func (gs *GTPServer) Serve() error {
	ctx := context.Background()
	if err := gs.srvConn.ListenAndServe(ctx); err != nil {
		fmt.Printf("GTP Server ListenAndServe Error: %v", err)
		return err
	}
	return nil
}

// Stop GTP-U server & close connections
func (gs *GTPServer) Stop() {
	gs.srvConn.Close()
}
