package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/divanov-web/shorturl/internal/grpcserver/pb"
	"github.com/divanov-web/shorturl/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedShortenerServer
	svc *service.URLService
}

// New конструктор Server
func New(svc *service.URLService) *Server {
	return &Server{svc: svc}
}

// CreateShort — аналог POST /api/shorten
func (s *Server) CreateShort(ctx context.Context, req *pb.CreateShortRequest) (*pb.CreateShortResponse, error) {
	shortURL, err := s.svc.CreateShort(ctx, req.GetUserId(), req.GetUrl())
	if err == service.ErrAlreadyExists {
		resp := &pb.CreateShortResponse{}
		resp.SetShortUrl(shortURL)
		resp.SetConflict(true)
		return resp, nil
	}
	if err != nil {
		return nil, fmt.Errorf("create short: %w", err)
	}
	resp := &pb.CreateShortResponse{}
	resp.SetShortUrl(shortURL)
	resp.SetConflict(false)
	return resp, nil
}

// ResolveShort — аналог GET /{id}
func (s *Server) ResolveShort(ctx context.Context, req *pb.ResolveShortRequest) (*pb.ResolveShortResponse, error) {
	orig, ok := s.svc.ResolveShort(ctx, req.GetId())
	resp := &pb.ResolveShortResponse{}
	resp.SetOriginalUrl(orig)
	resp.SetFound(ok)
	return resp, nil
}

// CreateShortBatch — аналог POST /api/shorten/batch
func (s *Server) CreateShortBatch(ctx context.Context, req *pb.CreateShortBatchRequest) (*pb.CreateShortBatchResponse, error) {
	items := req.GetItems()
	in := make([]service.BatchRequestItem, 0, len(items))
	for _, it := range items {
		in = append(in, service.BatchRequestItem{
			CorrelationID: it.GetCorrelationId(),
			OriginalURL:   it.GetOriginalUrl(),
		})
	}
	results, err := s.svc.CreateShortBatch(ctx, req.GetUserId(), in)
	if err != nil {
		return nil, fmt.Errorf("batch: %w", err)
	}
	out := make([]*pb.ShortBatchResult, 0, len(results))
	for _, r := range results {
		item := &pb.ShortBatchResult{}
		item.SetCorrelationId(r.CorrelationID)
		item.SetShortUrl(r.ShortURL)
		out = append(out, item)
	}
	resp := &pb.CreateShortBatchResponse{}
	resp.SetResults(out)
	return resp, nil
}

// Stats — аналог GET /api/internal/stats
func (s *Server) Stats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsResponse, error) {
	st, err := s.svc.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("stats: %w", err)
	}
	resp := &pb.StatsResponse{}
	resp.SetUrls(int32(st.URLs))
	resp.SetUsers(int32(st.Users))
	return resp, nil
}

// Runner — хелпер для запуска gRPC сервера.
type Runner struct {
	grpcServer *grpc.Server
	lis        net.Listener
}

// NewRunner — конструктор Runner
func NewRunner(addr string, svc *service.URLService, opts ...grpc.ServerOption) (*Runner, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", addr, err)
	}
	s := grpc.NewServer(opts...)
	pb.RegisterShortenerServer(s, New(svc))

	// Включаем серверную рефлексию для grpcurl list/describe
	reflection.Register(s)

	return &Runner{grpcServer: s, lis: lis}, nil
}

// ServeAsync — запуск gRPC сервера в отдельной горутине
func (r *Runner) ServeAsync() {
	go func() { _ = r.grpcServer.Serve(r.lis) }()
}

// GracefulStop — остановка gRPC сервера с ожиданием завершения текущих запросов
func (r *Runner) GracefulStop() {
	r.grpcServer.GracefulStop()
}
