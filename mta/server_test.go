package mta

import (
	"context"
	"gd92"
	"testing"
	"time"
)

func TestServerClientExchange(t *testing.T) {
	serverCfg := DefaultConfig()
	serverCfg.AckType = AckSlave
	serverCfg.VerifyPeriod = 1 * time.Hour

	srv, err := Listen("127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start server in background
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.Serve(ctx)
	}()

	// Connect client
	clientCfg := DefaultConfig()
	clientCfg.AckType = AckMaster
	clientCfg.VerifyPeriod = 1 * time.Hour

	cl, err := Dial(ctx, srv.Addr().String(), clientCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	// Build a mobilise command message
	msgBody := &gd92.MobiliseCommand{OpPeripherals: 0x0003, ManAckReq: true}
	msgData, err := msgBody.MarshalGD92()
	if err != nil {
		t.Fatal(err)
	}

	env := &gd92.Envelope{
		Source:       gd92.CommsAddress{Brigade: 1, Node: 0, Port: 1},
		Destinations: []gd92.CommsAddress{{Brigade: 1, Node: 10, Port: 4}},
		Priority:     1,
		ProtVers:     1,
		AckReq:       true,
		Seq:          1,
		MessageType:  gd92.MsgMobiliseCommand,
		Message:      msgData,
	}

	// Send from client
	if err := cl.Send(ctx, env); err != nil {
		t.Fatalf("client send: %v", err)
	}

	// Receive on server
	select {
	case rx := <-srv.Received():
		if rx.Err != nil {
			t.Fatalf("server receive error: %v", rx.Err)
		}
		if rx.Envelope.MessageType != gd92.MsgMobiliseCommand {
			t.Fatalf("expected MobiliseCommand, got type %d", rx.Envelope.MessageType)
		}
		mc, ok := rx.Message.(*gd92.MobiliseCommand)
		if !ok {
			t.Fatalf("expected *MobiliseCommand, got %T", rx.Message)
		}
		if mc.OpPeripherals != 0x0003 {
			t.Fatalf("OpPeripherals: expected 0x0003, got 0x%04x", mc.OpPeripherals)
		}

		// Send ACK reply from server to client
		ackData, _ := (&gd92.ACK{}).MarshalGD92()
		ackEnv := &gd92.Envelope{
			Source:       rx.Envelope.Destinations[0],
			Destinations: []gd92.CommsAddress{rx.Envelope.Source},
			Priority:     1,
			ProtVers:     1,
			Seq:          rx.Envelope.Seq,
			MessageType:  gd92.MsgACK,
			Message:      ackData,
		}
		if err := rx.Conn.SendEnvelope(ctx, ackEnv); err != nil {
			t.Fatalf("server send ACK: %v", err)
		}

	case <-ctx.Done():
		t.Fatal("timeout waiting for server to receive")
	}

	// Receive ACK on client
	select {
	case rx := <-cl.Received():
		if rx.Err != nil {
			t.Fatalf("client receive error: %v", rx.Err)
		}
		if rx.Envelope.MessageType != gd92.MsgACK {
			t.Fatalf("expected ACK, got type %d", rx.Envelope.MessageType)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for client to receive ACK")
	}

	cancel()
}

func TestServerMultipleMessages(t *testing.T) {
	serverCfg := DefaultConfig()
	serverCfg.AckType = AckSlave
	serverCfg.VerifyPeriod = 1 * time.Hour

	srv, err := Listen("127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go srv.Serve(ctx)

	clientCfg := DefaultConfig()
	clientCfg.AckType = AckMaster
	clientCfg.VerifyPeriod = 1 * time.Hour

	cl, err := Dial(ctx, srv.Addr().String(), clientCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	// Send multiple messages
	for i := 0; i < 5; i++ {
		msgBody := &gd92.ResourceStatus{
			Resources: []gd92.ResourceEntry{
				{
					Callsign:   "E21",
					AVLType:    0,
					StatusCode: uint8(i),
					Remarks:    "TEST",
				},
			},
		}
		msgData, _ := msgBody.MarshalGD92()

		env := &gd92.Envelope{
			Source:       gd92.CommsAddress{Brigade: 1, Node: 0, Port: 1},
			Destinations: []gd92.CommsAddress{{Brigade: 1, Node: 10, Port: 4}},
			Priority:     3,
			ProtVers:     1,
			Seq:          uint16(i),
			MessageType:  gd92.MsgResourceStatus,
			Message:      msgData,
		}

		if err := cl.Send(ctx, env); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	// Receive all 5
	for i := 0; i < 5; i++ {
		select {
		case rx := <-srv.Received():
			if rx.Err != nil {
				t.Fatalf("receive %d: %v", i, rx.Err)
			}
			if rx.Envelope.Seq != uint16(i) {
				t.Fatalf("receive %d: expected seq %d, got %d", i, i, rx.Envelope.Seq)
			}
		case <-ctx.Done():
			t.Fatalf("timeout waiting for message %d", i)
		}
	}

	cancel()
}
