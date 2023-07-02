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
		s := grpc.NewServer()
		agentV1.RegisterDistributeServiceServer(s, &distributeService{})
		s.Serve(c) // 在这里实现一个net.listener

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch

		return nil
	},
}

type distributeService struct {
	agentV1.UnimplementedDistributeServiceServer
}

var id = rand.Int63()

func (d *distributeService) AgentInfo(ctx context.Context, request *agentV1.AgentInfoRequest) (*agentV1.AgentInfoResponse, error) {
	return &agentV1.AgentInfoResponse{
		AgentID: fmt.Sprintf("%d", id),
	}, nil
}

func (d *distributeService) AgentStatus(ctx context.Context, request *agentV1.AgentStatusRequest) (*agentV1.AgentStatusResponse, error) {
	//TODO implement me
	panic("implement me")
}
