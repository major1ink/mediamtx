package RMS

import (
	"context"
	"fmt"

	"github.com/bluenviron/mediamtx/internal/conf"
	pb "github.com/major1ink/repGrpc/pkg/repGrpc"
	"google.golang.org/grpc"
)
type GrpcClient struct {
	ctx context.Context
	Use bool
	Server string
	Client pb.RMSClient
	Conn *grpc.ClientConn
}
func CreateGrpcClient(ctx context.Context, cfg conf.GRPC) (GrpcClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.GrpcAddress, cfg.GrpcPort), grpc.WithInsecure())
	if err != nil {
		return GrpcClient{}, err
	}
	client := GrpcClient{
		ctx:    ctx,
		Use:    cfg.Use,
		Server: cfg.ServerName,
		Client: pb.NewRMSClient(conn),
		Conn:   conn,
	}
	return client, nil
}
func (c *GrpcClient) Post (attribute,query string) (*pb.AnswerInsert, error){
	r,err := c.Client.Post(c.ctx, &pb.Insert{	Server: c.Server,	Attribute: attribute,	Query: query,})
	if err != nil {
		return nil, err
	}
	return r, nil

}
func (c *GrpcClient) Close () {
c.Conn.Close()
}