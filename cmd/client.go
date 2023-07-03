package cmd

import (
	"context"
	"fmt"
	"github.com/asjdf/smux"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	agentV1 "grpc-reverse-test/gen/proto/agent/v1"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Client此时实际上是提供服务的server端
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a client",
	RunE: func(cmd *cobra.Command, args []string) error {
		go func() {
			for {
				func() {
					// 工作流程 grpcTun 负责创建 net.listener
					// server 端收到建立链接后 执行 On 事件，由调用者管理会话
					conn, err := net.Dial("tcp", "localhost:5325")
					if err != nil {
						return
					}

					muxConf := smux.DefaultConfig()
					muxConf.Version = 2
					mux, err := smux.Client(conn, muxConf)
					if err != nil {
						return
					}

					var wg sync.WaitGroup
					wg.Add(1)
					go func() {
						s := grpc.NewServer()
						agentV1.RegisterAgentServiceServer(s, &agentService{})
						_ = s.Serve(mux)
						wg.Done()
					}()

					wg.Add(1)
					go func() {
						fmt.Println("start open stream")
						connToBackend, err := mux.OpenStream()
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
						fmt.Println("start auth")
						be := agentV1.NewBackendServiceClient(grpcConn)
						for {
							if _, err := be.AgentAuth(context.Background(),
								&agentV1.AgentAuthRequest{AgentID: fmt.Sprintf("%d", id), Token: "demoToken"}); err == nil {
								break
							} else {
								fmt.Println(err)
							}
							time.Sleep(time.Second)
						}
						wg.Done()
					}()
					wg.Wait()
				}()

				fmt.Println("conn break")
				time.Sleep(time.Second) // todo: 指数退避 demo不做实现
				fmt.Println("try reconnect")
			}
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
