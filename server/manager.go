package server

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	agentV1 "grpc-reverse-test/gen/proto/agent/v1"
	"grpc-reverse-test/pkg/grpcTun"
	"grpc-reverse-test/pkg/syncx"
	"net"
)

type AgentManager interface {
	Listen(listener net.Listener) AgentManager
}

type Agent interface {
	agentV1.DistributeServiceClient
}

type agent struct {
	conn *grpc.ClientConn

	agentV1.DistributeServiceClient
}

func (a *agent) Init() {
	a.DistributeServiceClient = agentV1.NewDistributeServiceClient(a.conn)
}

type stdAgentManager struct {
	tun *grpcTun.Server

	agent syncx.Map[string, Agent]
}

func (m *stdAgentManager) Listen(listener net.Listener) AgentManager {
	m.tun = grpcTun.NewServer().Listen(listener)
	return m
}

func (m *stdAgentManager) Init() AgentManager {
	m.tun.OnConnect(func(conn net.Conn) error {
		grpcConn, err := grpc.Dial("",
			grpc.WithInsecure(),
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return conn, nil
			}),
		)
		if err != nil {
			return err
		}

		client := agentV1.NewDistributeServiceClient(grpcConn)
		// todo: add some client verify process
		agentInfo, err := client.AgentInfo(context.Background(), &agentV1.AgentInfoRequest{})
		if err != nil {
			fmt.Printf("get agent info error: %v\n", err)
			return err
		} else {
			fmt.Printf("get agent info: id = %s\n", agentInfo.AgentID)
			a := &agent{conn: grpcConn}
			a.Init()
			m.agent.Store(agentInfo.AgentID, a)
		}
		return nil
	})
	return m
}

func (m *stdAgentManager) List() (list []Agent) {
	// todo: agent存活检测
	m.agent.Range(func(key string, value Agent) bool {
		list = append(list, value)
		return true
	})
	return
}

func (m *stdAgentManager) Get(agentID string) (Agent, bool) {
	// todo: agent存活检测
	return m.agent.Load(agentID)
}

func NewAgentManager() AgentManager {
	return &stdAgentManager{}
}
