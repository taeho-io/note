package server

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/taeho-io/auth"
	"github.com/taeho-io/note"
	"github.com/taeho-io/note/server/handler"
	"github.com/taeho-io/taeho-go/id"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type NoteServer struct {
	note.NoteServer

	cfg     Config
	db      *sql.DB
	id      id.ID
	authCli auth.AuthClient
}

func New(cfg Config) (*NoteServer, error) {
	dsn := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.Settings().PostgresHost,
		cfg.Settings().PostgresDBName,
		cfg.Settings().PostgresUser,
		cfg.Settings().PostgresPassword,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	for {
		err = db.Ping()
		if err != nil {
			logrus.Error(errors.Wrap(err, "db ping failed"))
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	tid := id.New()

	authCli := auth.GetAuthClient()

	return &NoteServer{
		cfg:     cfg,
		db:      db,
		id:      tid,
		authCli: authCli,
	}, nil
}

func Mock() *NoteServer {
	s, _ := New(MockConfig())
	return s
}

func (s *NoteServer) Config() Config {
	return s.cfg
}

func (s *NoteServer) DB() *sql.DB {
	return s.db
}

func (s *NoteServer) ID() id.ID {
	return s.id
}

func (s *NoteServer) AuthClient() auth.AuthClient {
	return s.authCli
}

func (s *NoteServer) RegisterServer(srv *grpc.Server) {
	note.RegisterNoteServer(srv, s)
}

func (s *NoteServer) Create(ctx context.Context, req *note.CreateRequest) (*note.CreateResponse, error) {
	return handler.Create(s.DB(), s.ID())(ctx, req)
}

func (s *NoteServer) Get(ctx context.Context, req *note.GetRequest) (*note.GetResponse, error) {
	return handler.Get(s.DB())(ctx, req)
}

func (s *NoteServer) List(ctx context.Context, req *note.ListRequest) (*note.ListResponse, error) {
	return handler.List(s.DB())(ctx, req)
}

func (s *NoteServer) Delete(ctx context.Context, req *note.DeleteRequest) (*note.DeleteResponse, error) {
	return handler.Delete(s.DB())(ctx, req)
}

func (s *NoteServer) Update(ctx context.Context, req *note.UpdateRequest) (*note.UpdateResponse, error) {
	return handler.Update(s.DB())(ctx, req)
}

func Serve(addr string, cfg Config) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logrusEntry := logrus.NewEntry(logrus.StandardLogger())

	grpcServer := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(
				grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			auth.TokenUnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logrusEntry),
			grpc_recovery.UnaryServerInterceptor(),
		),
	)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	noteServer, err := New(cfg)
	if err != nil {
		return err
	}
	note.RegisterNoteServer(grpcServer, noteServer)

	reflection.Register(grpcServer)
	return grpcServer.Serve(lis)
}
