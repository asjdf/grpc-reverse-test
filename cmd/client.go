package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	agentV1 "grpc-reverse-test/gen/proto/agent/v1"
	"grpc-reverse-test/pkg/grpcTun"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Client此时实际上是提供服务的server端
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a client",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 工作流程 grpcTun 负责创建 net.listener
		// server 端收到建立链接后 执行 On 事件，由调用者管理会话
		conn, err := net.Dial("tcp", "localhost:5325")
		if err != nil {
			return err
		}
		c := grpcTun.NewClient().Conn(conn)
		err = c.Init()
		if err != nil {
			return err
		}

		go func() {
			s := grpc.NewServer()
			agentV1.RegisterAgentServiceServer(s, &agentService{})
			s.Serve(c) // 在这里实现一个net.listener
		}()

		go func() {
			fmt.Println("start open stream")
			connToBackend, err := c.OpenStream()
			if err != nil {
				fmt.Printf("open conn to server error: %v\n", err)
				return
			}
			fmt.Println("start dial")
			grpcConn, err := grpc.Dial("",
				grpc.WithInsecure(),
				grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
					return connToBackend, nil
				}),
			)
			be := agentV1.NewBackendServiceClient(grpcConn)
			for {
				if _, err := be.AgentAuth(context.Background(),
					&agentV1.AgentAuthRequest{AgentID: fmt.Sprintf("%d", id), Token: "demoToken"}); err == nil {
					break
				}
				time.Sleep(time.Second)
			}

			fmt.Println("auth finished")
		}()

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch

		return nil
	},
}

type agentService struct {
	agentV1.UnimplementedAgentServiceServer
}

var id = rand.Int63()

func (d *agentService) AgentInfo(ctx context.Context, request *agentV1.AgentInfoRequest) (*agentV1.AgentInfoResponse, error) {
	return &agentV1.AgentInfoResponse{
		AgentID: fmt.Sprintf("%d", id),
	}, nil
}

func (d *agentService) AgentStatus(ctx context.Context, request *agentV1.AgentStatusRequest) (*agentV1.AgentStatusResponse, error) {
	//TODO implement me
	panic("implement me")
}
