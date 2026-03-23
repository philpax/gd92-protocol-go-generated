package gd92_test

import (
	"context"
	"fmt"
	"gd92"
	"gd92/mta"
	"time"
)

func ExampleCompress() {
	input := []byte("     HIGH STREET")
	compressed := gd92.Compress(input)
	decompressed, _ := gd92.Decompress(compressed)
	fmt.Printf("Original:     %q (%d bytes)\n", string(input), len(input))
	fmt.Printf("Compressed:   %d bytes\n", len(compressed))
	fmt.Printf("Decompressed: %q\n", decompressed)
	// Output:
	// Original:     "     HIGH STREET" (16 bytes)
	// Compressed:   14 bytes
	// Decompressed: "     HIGH STREET"
}

func ExampleCommsAddress() {
	addr := gd92.CommsAddress{Brigade: 5, Node: 42, Port: 3}
	fmt.Println(addr.String())

	// Encode to 3 bytes
	b := addr.MarshalAddress()
	fmt.Printf("Encoded: [%d %d %d]\n", b[0], b[1], b[2])

	// Decode back
	decoded := gd92.UnmarshalAddress(b)
	fmt.Printf("Decoded: Brigade=%d Node=%d Port=%d\n",
		decoded.Brigade, decoded.Node, decoded.Port)
	// Output:
	// B5/N42/P3
	// Encoded: [5 10 131]
	// Decoded: Brigade=5 Node=42 Port=3
}

func ExampleEnvelope() {
	// Create and encode a Mobilise_command message
	msg := &gd92.MobiliseCommand{
		OpPeripherals: 0x0003, // Station sounders + lights
		ManAckReq:     true,
	}
	msgData, _ := msg.MarshalGD92()

	// Wrap in an envelope
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

	// Marshal to wire format (includes BCC)
	wireData, _ := env.MarshalEnvelope()
	fmt.Printf("Envelope: %d bytes\n", len(wireData))

	// Unmarshal back
	decoded, _ := gd92.UnmarshalEnvelope(wireData)
	fmt.Printf("Source: %s\n", decoded.Source.String())
	fmt.Printf("Dest:   %s\n", decoded.Destinations[0].String())
	fmt.Printf("Type:   %d (MobiliseCommand)\n", decoded.MessageType)

	// Parse the message body
	parsedMsg, _ := gd92.ParseEnvelopeMessage(decoded)
	mc := parsedMsg.(*gd92.MobiliseCommand)
	fmt.Printf("OpPeripherals: 0x%04x\n", mc.OpPeripherals)
	fmt.Printf("ManAckReq: %v\n", mc.ManAckReq)
	// Output:
	// Envelope: 16 bytes
	// Source: B1/N0/P1
	// Dest:   B1/N10/P4
	// Type:   1 (MobiliseCommand)
	// OpPeripherals: 0x0003
	// ManAckReq: true
}

func ExampleServer() {
	// Start a server
	cfg := mta.DefaultConfig()
	cfg.AckType = mta.AckSlave
	cfg.VerifyPeriod = 1 * time.Hour

	srv, err := mta.Listen("127.0.0.1:0", cfg)
	if err != nil {
		panic(err)
	}
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go srv.Serve(ctx)

	// Connect a client
	clientCfg := mta.DefaultConfig()
	clientCfg.AckType = mta.AckMaster
	clientCfg.VerifyPeriod = 1 * time.Hour

	cl, err := mta.Dial(ctx, srv.Addr().String(), clientCfg)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	// Send a resource status
	msg := &gd92.ResourceStatus{
		Resources: []gd92.ResourceEntry{{
			Callsign:   "E21",
			StatusCode: 4, // Available at Base
		}},
	}
	msgData, _ := msg.MarshalGD92()

	env := &gd92.Envelope{
		Source:       gd92.CommsAddress{Brigade: 1, Node: 5, Port: 1},
		Destinations: []gd92.CommsAddress{{Brigade: 1, Node: 0, Port: 1}},
		Priority:     3,
		ProtVers:     1,
		Seq:          1,
		MessageType:  gd92.MsgResourceStatus,
		Message:      msgData,
	}

	if err := cl.Send(ctx, env); err != nil {
		panic(err)
	}

	// Receive on server
	rx := <-srv.Received()
	rs := rx.Message.(*gd92.ResourceStatus)
	fmt.Printf("Received: %s status=%d\n", rs.Resources[0].Callsign, rs.Resources[0].StatusCode)
	cancel()
	// Output:
	// Received: E21 status=4
}
