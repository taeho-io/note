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

type CreateHandlerFunc func(ctx context.Context, request *note.CreateRequest) (*note.CreateResponse, error)

func Create(db *sql.DB, tid id.ID) CreateHandlerFunc {
	return func(ctx context.Context, req *note.CreateRequest) (*note.CreateResponse, error) {
		claims, err := token.GetClaimsFromMD(ctx)
		if err != nil {
			return nil, ErrInvalidToken
		}

		noteID := tid.Must()
		userID := claims.UserID

		n := &models.Note{
			ID:        noteID,
			CreatedBy: userID,
			Title:     req.Title,
			BodyType:  req.BodyType.String(),
			Body:      req.Body,
		}

		err = n.Insert(ctx, db, boil.Infer())

		if err != nil {
			return nil, err
		}

		return &note.CreateResponse{
			NoteId: n.ID,
		}, nil
	}
}
