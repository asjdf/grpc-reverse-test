package manage

import (
	"context"
	"fmt"
	"github.com/xtaci/smux"
	"google.golang.org/grpc"
	agentV1 "grpc-reverse-test/gen/proto/agent/v1"
	"grpc-reverse-test/pkg/syncx"
	"net"
)

type AgentManager interface {
	Listen(listener net.Listener) AgentManager
	Serve() AgentManager
	// List all agents
	List() (list []Agent)
	// Get agent by agentID
	Get(agentID string) (Agent, bool)
	// RegAgent add agent to manager
	RegAgent(agentID string, a Agent)
}

func NewAgentManager() AgentManager {
	return &stdAgentManager{}
}

type stdAgentManager struct {
	listener net.Listener

	agent syncx.Map[string, Agent]
}

func (m *stdAgentManager) Listen(listener net.Listener) AgentManager {
	m.listener = listener
	return m
}

func (m *stdAgentManager) Serve() AgentManager {
	go func() {
		for {
			conn, err := m.listener.Accept()
			if err != nil {
				fmt.Printf("accept error: %v\n", err)
				continue
			}

			a := &agent{}
			muxConf := smux.DefaultConfig()
			muxConf.Version = 2
			a.mux, err = smux.Server(conn, muxConf)
			if err != nil {
				fmt.Printf("init smux error: %v\n", err)
				continue
			}

			go func() {
				connToClient, err := a.mux.OpenStream()
				if err != nil {
					fmt.Printf("open conn to client error: %v\n", err)
					return
				}
				grpcConn, err := grpc.Dial("",
					grpc.WithInsecure(),
					grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
						return connToClient, nil
					}),
				)
				if err != nil {
					return
				}
				client := agentV1.NewAgentServiceClient(grpcConn)
				agentInfo, err := client.AgentInfo(context.Background(), &agentV1.AgentInfoRequest{})
				if err != nil {
					fmt.Printf("get agent info error: %v\n", err)
					return
				} else {
					fmt.Printf("agent connected: id = %s\n", agentInfo.AgentID)
				}
			}()

			go func() {
				s := grpc.NewServer()
				agentV1.RegisterBackendServiceServer(s, &agentAuthService{manager: m, agent: a})
				s.Serve(&smuxWarp{Session: a.mux})
			}()
		}
	}()
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

func (m *stdAgentManager) RegAgent(agentID string, a Agent) {
	m.agent.Store(agentID, a)
}

type agentAuthService struct {
	manager AgentManager
	agent   Agent

	agentV1.UnimplementedBackendServiceServer
}

func (a *agentAuthService) AgentAuth(ctx context.Context, request *agentV1.AgentAuthRequest) (*agentV1.AgentAuthResponse, error) {
	if request.Token == "demoToken" {
		fmt.Println("auth seccess")
		a.agent.Init()
		a.manager.RegAgent(request.AgentID, a.agent)
	} else {
		fmt.Println("auth failed")
		a.agent.Disconnect()
	}
	return &agentV1.AgentAuthResponse{Success: true}, nil
}

type smuxWarp struct {
	*smux.Session
}

func (s *smuxWarp) Accept() (net.Conn, error) {
	return s.Session.AcceptStream()
}

func (s *smuxWarp) Addr() net.Addr {
	return s.Session.LocalAddr()
}
