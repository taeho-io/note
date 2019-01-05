package handler

import (
	"database/sql"

	"github.com/taeho-io/auth/pkg/token"
	"github.com/taeho-io/note"
	"github.com/taeho-io/note/server/models"
	"github.com/taeho-io/taeho-go/id"
	"github.com/volatiletech/sqlboiler/boil"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidToken = status.Error(codes.Unauthenticated, "invalid token")
)

type CreateHandlerFunc func(ctx context.Context, req *note.CreateRequest) (*note.CreateResponse, error)

func Create(db *sql.DB, tid id.ID) CreateHandlerFunc {
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

		noteID := tid.Must()

		n := &models.Note{
			ID:        noteID,
			CreatedBy: req.CreatedBy,
			Title:     req.Title,
			BodyType:  req.BodyType.String(),
			Body:      req.Body,
		}

		if err := n.Insert(ctx, db, boil.Infer()); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &note.CreateResponse{
			NoteId: n.ID,
		}, nil
	}
}
