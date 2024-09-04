package repgrpc

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/major1ink/repGrpc/pkg/repGrpc"
	"google.golang.org/grpc"
)

type MockGrpcClient struct {
	ctrl     *gomock.Controller
	recorder *MockGrpcClientMockRecorder
}

// MockGrpcClientMockRecorder is the mock recorder for MockGrpcClient
type MockGrpcClientMockRecorder struct {
	mock *MockGrpcClient
}

// NewMockGrpcClient creates a new mock instance
func NewMockGrpcClient(ctrl *gomock.Controller) *MockGrpcClient {
	mock := &MockGrpcClient{ctrl: ctrl}
	mock.recorder = &MockGrpcClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockGrpcClient) EXPECT() *MockGrpcClientMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockGrpcClient) Get(ctx context.Context, in *pb.Select, opts ...grpc.CallOption) (*pb.AnswerSelect, error) {
	// m.ctrl.T.Helper()
	// ret := m.ctrl.Call(m, "Get", ctx, in)
	// ret0, _ := ret[0].(*pb.AnswerSelect)
	// ret1, _ := ret[1].(error)
	return nil, nil
}

// Get indicates an expected call of Get
func (mr *MockGrpcClientMockRecorder) Get(ctx, in interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockGrpcClient)(nil).Get), ctx, in)
}


func (mr *MockGrpcClientMockRecorder) Post(ctx, in interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Post", reflect.TypeOf((*MockGrpcClient)(nil).Post), ctx, in)
}


func (m *MockGrpcClient) Post(ctx context.Context, in *pb.Insert, opts ...grpc.CallOption) (*empty.Empty, error) {
    	m.ctrl.T.Helper()
	// ret := m.ctrl.Call(m, "Post", ctx, in)
	// ret0, _ := ret[0].(*empty.Empty)
	// ret1, _ := ret[1].(error)
	return nil, nil
}



