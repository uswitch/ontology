package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"
)

type Origin int

const (
	ServerOrigin = Origin(iota)
	ClientOrigin
)

type BufferMessage struct {
	typ int
	op  OperationMessage
}

type LogMessage struct {
	origin Origin
	msg    *BufferMessage
}

type ReadWriterBuffer struct {
	in  chan *BufferMessage
	out chan *BufferMessage
}

func (buf *ReadWriterBuffer) ReadMessage() (int, []byte, error) {
	msg := <-buf.in
	if msg == nil {
		return 0, nil, fmt.Errorf("No more data")
	}

	data, err := json.Marshal(msg.op)

	return msg.typ, data, err
}

func (buf *ReadWriterBuffer) WriteMessage(typ int, data []byte) error {
	var om OperationMessage

	err := json.Unmarshal(data, &om)
	if err != nil {
		return err
	}

	buf.out <- &BufferMessage{typ, om}
	return nil
}

type Channel struct {
	Client *ReadWriterBuffer
	Server *ReadWriterBuffer

	Log []*LogMessage

	fromClient chan *BufferMessage
	toServer   chan *BufferMessage
	fromServer chan *BufferMessage
	toClient   chan *BufferMessage
}

func NewChannel(queueSize int) *Channel {
	fromClient := make(chan *BufferMessage, queueSize)
	toServer := make(chan *BufferMessage, queueSize)
	fromServer := make(chan *BufferMessage, queueSize)
	toClient := make(chan *BufferMessage, queueSize)

	channel := &Channel{
		Client: &ReadWriterBuffer{in: toClient, out: fromClient},
		Server: &ReadWriterBuffer{in: toServer, out: fromServer},

		fromClient: fromClient,
		toServer:   toServer,
		fromServer: fromServer,
		toClient:   toClient,
	}

	go func() {
		for {
			select {
			case m := <-fromClient:
				if m == nil {
					return
				}
				log.Printf("[CLIENT] %v", m.op)
				channel.Log = append(channel.Log, &LogMessage{ClientOrigin, m})
				toServer <- m

			case m := <-fromServer:
				if m == nil {
					return
				}
				log.Printf("[SERVER] %v", m.op)
				channel.Log = append(channel.Log, &LogMessage{ServerOrigin, m})
				toClient <- m
			}
		}
	}()

	return channel
}

func (c *Channel) Close() {
	close(c.fromClient)
	close(c.toClient)
	close(c.fromServer)
	close(c.toServer)
}

func TestProtocol(t *testing.T) {
	ch := NewChannel(1)
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := NewClientChannel(ctx, ch.Client)
	server := NewServerChannel(ctx, ch.Server,
		func(ctx context.Context, operation OperationParams) (chan *OperationResult, error) {
			ch := make(chan *OperationResult, 1)

			ch <- &OperationResult{}

			close(ch)

			return ch, nil
		},
	)

	serverDone := make(chan bool)

	go func() {
		err := server.Accept(ctx)
		if err != nil {
			t.Fatalf("Error from server.Accept(): %v", err)
		}

		server.Listen(ctx)

		serverDone <- true
	}()

	err := client.Connect(ctx, ConnectionParams{})
	if err != nil {
		t.Fatalf("Error from client.Connect(): %v", err)
	}

	opCh, err := client.Operation(ctx, OperationParams{
		Query: "wibble",
	})
	if err != nil {
		t.Fatalf("Error from client.Operation(): %v", err)
	}

	dataCtx, _ := context.WithTimeout(ctx, 100*time.Millisecond)

	select {
	case <-opCh:
	case <-dataCtx.Done():
		t.Errorf("Didn't get back any data")
	}

	err = client.Close(ctx)
	if err != nil {
		t.Fatalf("Error from client.Close(): %v", err)
	}

	<-serverDone

	expectedSequence := []struct {
		origin Origin
		typ    MessageType
	}{
		{ClientOrigin, GQL_CONNECTION_INIT},
		{ServerOrigin, GQL_CONNECTION_ACK},
		{ClientOrigin, GQL_START},
		{ServerOrigin, GQL_DATA},
		{ServerOrigin, GQL_COMPLETE},
		{ClientOrigin, GQL_CONNECTION_TERMINATE},
	}

	if len(ch.Log) != len(expectedSequence) {
		t.Errorf("The number of expected messages %d != actual messages %d", len(expectedSequence), len(ch.Log))
	}

	for idx, log := range ch.Log {
		if log.origin != expectedSequence[idx].origin {
			t.Errorf("[%d] origins didn't match", idx)
		}

		if log.msg.op.Type != expectedSequence[idx].typ {
			t.Errorf("[%d] types didin't match. %s != %s", idx, log.msg.op.Type, expectedSequence[idx].typ)
		}
	}
}
