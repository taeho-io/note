package handler

import (
	"database/sql"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/taeho-io/auth/pkg/token"
	"github.com/taeho-io/idl/gen/go/note"
	"github.com/taeho-io/note/server/models"
	"github.com/volatiletech/sqlboiler/boil"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	idxNoteUniqueViolation = `duplicate key value violates unique constraint "note_pkey"`
)

var (
	ErrNoteAlreadyExists = status.Error(codes.AlreadyExists, "note already exists")
	ErrInvalidToken      = status.Error(codes.Unauthenticated, "invalid token")
)

type CreateHandlerFunc func(ctx context.Context, req *note.CreateRequest) (*note.CreateResponse, error)

func Create(db *sql.DB) CreateHandlerFunc {
	return func(ctx context.Context, req *note.CreateRequest) (*note.CreateResponse, error) {
		claims, err := token.GetClaimsFromMD(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		userID := claims.UserID

		if req.CreatedBy != userID {
			return nil, ErrInvalidToken
		}

		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		n := &models.Note{
			ID:        req.NoteId,
			CreatedBy: req.CreatedBy,
			Title:     req.Title,
			BodyType:  req.BodyType.String(),
			Body:      req.Body,
		}

		createdAt, err := ptypes.Timestamp(req.CreatedAt)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		n.CreatedAt = createdAt

		updatedAt, err := ptypes.Timestamp(req.UpdatedAt)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		n.UpdatedAt = updatedAt

		if err := n.Insert(ctx, db, boil.Infer()); err != nil {
			if strings.Contains(err.Error(), idxNoteUniqueViolation) {
				return nil, ErrNoteAlreadyExists
			}
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &note.CreateResponse{}, nil
	}
}
