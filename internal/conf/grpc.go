package conf

type GRPC struct {
	Use bool `json:"use"`
	GrpcAddress string `json:"grpcAddress"`
	GrpcPort int `json:"grpcPort"`
	ServerName string `json:"serverName"`
	UseCodeMPAttribute bool `json:"useCodeMPAttribute"`
}

type LossСatcher struct {
	Use bool `json:"use"`
	GrpcAddress string `json:"grpcAddress"`
	GrpcPort int `json:"grpcPort"`
	ServerName string `json:"serverName"`
}

func (grpc *GRPC) setDefaultsRMS() {
	grpc.Use = false
	grpc.GrpcAddress = "127.0.0.1"
	grpc.GrpcPort = 8080
	grpc.ServerName = ""
	grpc.UseCodeMPAttribute=false
}

func (grpc *LossСatcher) setDefaultsLossСatcher() {
	grpc.Use = false
	grpc.GrpcAddress = "127.0.0.1"
	grpc.GrpcPort = 8085
	grpc.ServerName = ""
}