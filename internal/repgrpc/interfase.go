package repgrpc

import (
	pb "github.com/major1ink/repGrpc/pkg/repGrpc"
)
type Grpc interface {
	Post() (error)
	Select (streamName, argument string) (*pb.AnswerSelect, error)
}