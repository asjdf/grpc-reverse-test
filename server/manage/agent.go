package manage

import (
	"context"
	"fmt"
	"github.com/xtaci/smux"
	"google.golang.org/grpc"
	agentV1 "grpc-reverse-test/gen/proto/agent/v1"
	"net"
	"time"
)

type Agent interface {
	agentV1.AgentServiceClient

	ID() string
	LastSeen() time.Time
	Init() error // Init after auth passed
	Disconnect() error
}

type agent struct {
	id string

	manager  AgentManager
	mux      *smux.Session
	conn     *grpc.ClientConn
	lastSeen time.Time

	agentV1.AgentServiceClient
}

func (a *agent) ID() string {
	return a.id
}

func (a *agent) LastSeen() time.Time {
	return a.lastSeen
}

func (a *agent) Init() error {
	connToClient, err := a.mux.OpenStream()
	if err != nil {
		fmt.Printf("open conn to client error: %v\n", err)
		return err
	}
	a.conn, err = grpc.Dial("",
		grpc.WithChainUnaryInterceptor(a.GrpcUnaryClientInterceptor()),
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

// Disconnect will remove the agent from manager's
// agentPool disconnect the agent.
func (a *agent) Disconnect() error {
	a.manager.Remove(a.id)
	a.conn.Close()
	a.mux.Close()
	return nil
}

// GrpcUnaryClientInterceptor was used to update last seen time
// the last seen time will be refreshed when err is nil
func (a *agent) GrpcUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			a.lastSeen = time.Now()
		}
		return err
	}
}

// GrpcUnaryServerInterceptor was used to update last seen time
// the last seen time will be refreshed when err is nil
func (a *agent) GrpcUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		resp, err := handler(ctx, req)
		if err == nil {
			a.lastSeen = time.Now()
		}
		return resp, err
	}
}
