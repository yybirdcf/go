package usersvc

// This file provides server-side bindings for the gRPC transport.
// It utilizes the transport/grpc.Server.

import (
	stdopentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"

	"go/microservice/services/modules/usersvc/pb"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	"go/microservice/services/utils"
)

// MakeGRPCServer makes a set of endpoints available as a gRPC AddServer.
func MakeGRPCServer(ctx context.Context, endpoints Endpoints, tracer stdopentracing.Tracer, logger log.Logger) pb.UsersvcServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}
	return &grpcServer{
		ctx: ctx,
		getUserinfo: grpctransport.NewServer(
			ctx,
			endpoints.GetUserinfoEndpoint,
			DecodeGRPCGetUserinfoRequest,
			EncodeGRPCGetUserinfoResponse,
			append(options, grpctransport.ServerBefore(opentracing.FromGRPCRequest(tracer, "GetUserinfo", logger)))...,
		),
	}
}

type grpcServer struct {
	ctx         context.Context
	getUserinfo grpctransport.Handler
}

func (s *grpcServer) GetUserinfo(ctx context.Context, req *pb.GetUserinfoRequest) (*pb.GetUserinfoResponse, error) {
	_, rep, err := s.getUserinfo.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.GetUserinfoResponse), nil
}

// DecodeGRPCSumRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC sum request to a user-domain sum request. Primarily useful in a server.
func DecodeGRPCGetUserinfoRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req, ok := grpcReq.(*pb.GetUserinfoRequest)
	if !ok {
		return nil, utils.Str2Err("interface conversion failed")
	}
	return getUserinfoRequest{Id: int64(req.Id)}, nil
}

func EncodeGRPCGetUserinfoRequest(_ context.Context, req interface{}) (interface{}, error) {
	grpcReq, ok := req.(getUserinfoRequest)
	if !ok {
		return nil, utils.Str2Err("interface conversion failed")
	}
	return &pb.GetUserinfoRequest{Id: int64(grpcReq.Id)}, nil
}

// DecodeGRPCSumResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC sum reply to a user-domain sum response. Primarily useful in a client.
func DecodeGRPCGetUserinfoResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply, ok := grpcReply.(*pb.GetUserinfoResponse)
	if !ok {
		return nil, utils.Str2Err("interface conversion failed")
	}

	return getUserinfoResponse{V: Userinfo{
		Id:             reply.Userinfo.Id,
		Username:       reply.Userinfo.Username,
		Phone:          reply.Userinfo.Phone,
		Sex:            reply.Userinfo.Sex,
		Avatar:         reply.Userinfo.Avatar,
		Gouhao:         reply.Userinfo.Gouhao,
		Birthday:       reply.Userinfo.Birthday,
		Avatars:        reply.Userinfo.Avatars,
		Signature:      reply.Userinfo.Signature,
		Appfrom:        reply.Userinfo.Appfrom,
		Appver:         reply.Userinfo.Appver,
		BackgroudImage: reply.Userinfo.BackgroudImage,
		UpdateAppver:   reply.Userinfo.UpdateAppver,
		Privacy:        reply.Userinfo.Privacy,
		LoadRecTags:    reply.Userinfo.LoadRecTags,
		GamePower:      reply.Userinfo.GamePower,
		Mark:           reply.Userinfo.Mark,
		Level:          reply.Userinfo.Level,
		QuestionPhoto:  reply.Userinfo.QuestionPhoto,
		Lan:            reply.Userinfo.Lan,
		Notify:         reply.Userinfo.Notify,
		BindGameIds:    reply.Userinfo.BindGameIds,
		UserPosition:   reply.Userinfo.UserPosition,
		UserStatus:     reply.Userinfo.UserStatus,
		ImToken:        reply.Userinfo.ImToken,
	}, Err: utils.Str2Err(reply.Err)}, nil
}

func EncodeGRPCGetUserinfoResponse(_ context.Context, resp interface{}) (interface{}, error) {
	grpcReply, ok := resp.(getUserinfoResponse)
	if !ok {
		return nil, utils.Str2Err("interface conversion failed")
	}

	return &pb.GetUserinfoResponse{Userinfo: &pb.Userinfo{
		Id:             grpcReply.V.Id,
		Username:       grpcReply.V.Username,
		Phone:          grpcReply.V.Phone,
		Sex:            grpcReply.V.Sex,
		Avatar:         grpcReply.V.Avatar,
		Gouhao:         grpcReply.V.Gouhao,
		Birthday:       grpcReply.V.Birthday,
		Avatars:        grpcReply.V.Avatars,
		Signature:      grpcReply.V.Signature,
		Appfrom:        grpcReply.V.Appfrom,
		Appver:         grpcReply.V.Appver,
		BackgroudImage: grpcReply.V.BackgroudImage,
		UpdateAppver:   grpcReply.V.UpdateAppver,
		Privacy:        grpcReply.V.Privacy,
		LoadRecTags:    grpcReply.V.LoadRecTags,
		GamePower:      grpcReply.V.GamePower,
		Mark:           grpcReply.V.Mark,
		Level:          grpcReply.V.Level,
		QuestionPhoto:  grpcReply.V.QuestionPhoto,
		Lan:            grpcReply.V.Lan,
		Notify:         grpcReply.V.Notify,
		BindGameIds:    grpcReply.V.BindGameIds,
		UserPosition:   grpcReply.V.UserPosition,
		UserStatus:     grpcReply.V.UserStatus,
		ImToken:        grpcReply.V.ImToken,
	}, Err: utils.Err2Str(grpcReply.Err)}, nil
}
