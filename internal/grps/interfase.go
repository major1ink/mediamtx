package RMS

import pb "github.com/major1ink/repGrpc/pkg/repGrpc"
type Grpc interface {
	Post() (*pb.AnswerInsert, error)
}