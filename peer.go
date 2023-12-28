package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Peer interface {
	Send(b []byte) error
}

type WSPeer struct {
	conn   *websocket.Conn
	server *Server
}

func NewWSPeer(conn *websocket.Conn, s *Server) *WSPeer {
	p := &WSPeer{
		conn:   conn,
		server: s,
	}
	go p.readLoop()
	return p
}

func (p *WSPeer) readLoop() {
	var msg WSMessage
	for {
		if err := p.conn.ReadJSON(&msg); err != nil {
			fmt.Println("ws peer read error: ", err)
			return
		}
		if err := p.handleMessage(msg); err != nil {
			fmt.Println("ws handle message error: ", err)
			return
		}
	}
}

func (p *WSPeer) handleMessage(msg WSMessage) error {
	//need validation
	if msg.Action == "subscribe" {
		p.server.AddPeerToTopics(p, msg.Topics...)
	}
	return nil
}

func (p *WSPeer) Send(b []byte) error {
	return p.conn.WriteMessage(websocket.BinaryMessage, b)
}
