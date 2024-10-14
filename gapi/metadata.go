package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	GrpcgatewayUserAgent = "grpcgateway-user-agent" // http请求的key
	UserAgent            = "user-agent"             // grpc请求的key
	XForwardedFor        = "x-forwarded-for"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

// 从metadata中获取UserAgent和ClientIP
func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// 获取http请求的UserAgent
		if userAgents := md.Get(GrpcgatewayUserAgent); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		// 获取grpc请求的UserAgent
		if userAgents := md.Get(UserAgent); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		// 获取http请求的ClientIP
		if clientIPs := md.Get(XForwardedFor); len(clientIPs) > 0 {
			mtdt.ClientIP = clientIPs[0]
		}
	}
	// 获取grpc请求的ClientIP
	p, ok := peer.FromContext(ctx)
	if ok {
		mtdt.ClientIP = p.Addr.String()
	}

	return mtdt
}
