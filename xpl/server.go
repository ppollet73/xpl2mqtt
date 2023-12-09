package xpl

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Server struct {
	conn *net.UDPConn
	stop bool
	mqtt *mqtt.Client
}

var XPLPort = 3865

func NewServer(port int, client *mqtt.Client) *Server {
	listen, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("unable to resolve udp address: %s", err.Error())
	}
	conn, err := net.ListenUDP("udp4", listen)
	if err != nil {
		log.Fatalf("unable to initialize UDPConn: %s", err.Error())
	}
	return &Server{
		conn: conn,
		stop: false,
		mqtt: client,
	}
}

func (p *Server) Stop() {
	p.stop = true
}

func (p *Server) Run() error {
	buffer := make([]byte, 2048)
	for !p.stop {
		recvSize, _, err := p.conn.ReadFrom(buffer)
		if err != nil {
			slog.Debug("error receiving udp packet", "error", err.Error())
		}
		if recvSize == 2048 {
			slog.Warn("received udp message is the same size as the buffer, it may have been truncated")
		}

		pkt, err := DecodePacket(string(buffer[:recvSize]))
		if err != nil {
			slog.Warn("error decoding packet: " + err.Error())
		} else {
			ProcessXPL(&pkt, p.mqtt)
		}
	}
	return nil
}

func (p *Server) Write(pkt *XPLPacket, addr *net.UDPAddr, nbPackets int) error {
	for i := 0; i < nbPackets; i++ {
		_, err := p.conn.WriteTo([]byte(EncodePacket(*pkt)), addr)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}
