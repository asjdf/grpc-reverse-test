package manage

import (
	"context"
	"fmt"
	"github.com/xtaci/smux"
	"google.golang.org/grpc"
	agentV1 "grpc-reverse-test/gen/proto/agent/v1"
	"net"
)

type Agent interface {
	agentV1.AgentServiceClient

	Init() error // InitAfterAuthPassed
	Disconnect() error
}

type agent struct {
	mux  *smux.Session
	conn *grpc.ClientConn

	agentV1.AgentServiceClient
}

func (a *agent) Init() error {
	connToClient, err := a.mux.OpenStream()
	if err != nil {
		fmt.Printf("open conn to client error: %v\n", err)
		return err
	}
	a.conn, err = grpc.Dial("",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return connToClient, nil
		}),
	)
	if err != nil {
		return err
	}
	a.AgentServiceClient = agentV1.NewAgentServiceClient(a.conn)
	return nil
}

func (a *agent) Disconnect() error {
	a.conn.Close()
	a.mux.Close()
	return nil
}
