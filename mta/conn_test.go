package mta

import (
	"context"
	"gd92"
	"net"
	"testing"
	"time"
)

func TestConnSendReceiveEnvelope(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	masterCfg := DefaultConfig()
	masterCfg.AckType = AckMaster
	masterCfg.VerifyPeriod = 1 * time.Hour // disable for test
	masterCfg.FrameTimeout = 2 * time.Second

	slaveCfg := DefaultConfig()
	slaveCfg.AckType = AckSlave
	slaveCfg.VerifyPeriod = 1 * time.Hour
	slaveCfg.FrameTimeout = 2 * time.Second

	masterConn := NewConn(client, masterCfg)
	slaveConn := NewConn(server, slaveCfg)
	defer masterConn.Close()
	defer slaveConn.Close()

	env := &gd92.Envelope{
		Source:       gd92.CommsAddress{Brigade: 1, Node: 0, Port: 1},
		Destinations: []gd92.CommsAddress{{Brigade: 1, Node: 10, Port: 4}},
		Priority:     1,
		ProtVers:     1,
		AckReq:       true,
		Seq:          42,
		MessageType:  gd92.MsgACK,
		Message:      nil,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send from master, receive on slave
	errCh := make(chan error, 1)
	go func() {
		errCh <- masterConn.SendEnvelope(ctx, env)
	}()

	gotEnv, err := slaveConn.ReadEnvelope(ctx)
	if err != nil {
		t.Fatalf("slave ReadEnvelope: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("master SendEnvelope: %v", err)
	}

	if !gotEnv.Source.Equal(env.Source) {
		t.Fatalf("source mismatch: expected %v, got %v", env.Source, gotEnv.Source)
	}
	if gotEnv.Seq != 42 {
		t.Fatalf("seq: expected 42, got %d", gotEnv.Seq)
	}
	if gotEnv.MessageType != gd92.MsgACK {
		t.Fatalf("message type: expected %d, got %d", gd92.MsgACK, gotEnv.MessageType)
	}
}

func TestConnENQHandling(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	slaveCfg := DefaultConfig()
	slaveCfg.AckType = AckSlave
	slaveCfg.VerifyPeriod = 1 * time.Hour

	slaveConn := NewConn(server, slaveCfg)
	defer slaveConn.Close()

	// The readLoop goroutine started by NewConn handles ENQ.
	// Send ENQ directly from the "remote" side.
	client.Write([]byte{gd92.ENQ})

	// Read the ack response
	buf := make([]byte, 1)
	client.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("read ack: %v", err)
	}
	if n != 1 || buf[0] != gd92.ACKS {
		t.Fatalf("expected ACKS (0x07), got 0x%02x", buf[0])
	}
}

func TestConnBidirectional(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	masterCfg := DefaultConfig()
	masterCfg.AckType = AckMaster
	masterCfg.VerifyPeriod = 1 * time.Hour
	masterCfg.FrameTimeout = 2 * time.Second

	slaveCfg := DefaultConfig()
	slaveCfg.AckType = AckSlave
	slaveCfg.VerifyPeriod = 1 * time.Hour
	slaveCfg.FrameTimeout = 2 * time.Second

	masterConn := NewConn(client, masterCfg)
	slaveConn := NewConn(server, slaveCfg)
	defer masterConn.Close()
	defer slaveConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send from master to slave
	env1 := &gd92.Envelope{
		Source:       gd92.CommsAddress{Brigade: 1, Node: 0, Port: 1},
		Destinations: []gd92.CommsAddress{{Brigade: 2, Node: 0, Port: 1}},
		Priority:     1,
		ProtVers:     1,
		Seq:          1,
		MessageType:  gd92.MsgMobiliseCommand,
		Message:      []byte{0x00, 0x03, 0x01}, // op_peripherals=3, man_ack=true
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- masterConn.SendEnvelope(ctx, env1)
	}()

	got1, err := slaveConn.ReadEnvelope(ctx)
	if err != nil {
		t.Fatalf("slave read: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("master send: %v", err)
	}

	// Send reply from slave to master
	env2 := &gd92.Envelope{
		Source:       got1.Destinations[0],
		Destinations: []gd92.CommsAddress{got1.Source},
		Priority:     1,
		ProtVers:     1,
		Seq:          got1.Seq,
		MessageType:  gd92.MsgACK,
		Message:      nil,
	}

	go func() {
		errCh <- slaveConn.SendEnvelope(ctx, env2)
	}()

	got2, err := masterConn.ReadEnvelope(ctx)
	if err != nil {
		t.Fatalf("master read: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("slave send: %v", err)
	}

	if got2.MessageType != gd92.MsgACK {
		t.Fatalf("expected ACK, got type %d", got2.MessageType)
	}
}
