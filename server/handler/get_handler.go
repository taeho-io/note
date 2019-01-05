package handler

import (
	"database/sql"

	"github.com/taeho-io/auth/pkg/token"
	"github.com/taeho-io/note"
	"github.com/taeho-io/note/server/models"
	"github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNoteNotFound = status.Error(codes.NotFound, "note not found")
)

type GetHandlerFunc func(ctx context.Context, request *note.GetRequest) (*note.GetResponse, error)

func Get(db *sql.DB) GetHandlerFunc {
	return func(ctx context.Context, req *note.GetRequest) (*note.GetResponse, error) {
		claims, err := token.GetClaimsFromMD(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		userID := claims.UserID

		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		n, err := models.Notes(qm.Where("id=?", req.NoteId)).One(ctx, db)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return nil, ErrNoteNotFound
			}

			return nil, status.Error(codes.Internal, err.Error())
		}

		if n.CreatedBy != userID {
			return nil, ErrNoteNotFound
		}

		return &note.GetResponse{
			NoteId:    n.ID,
			CreatedBy: n.CreatedBy,
			Title:     n.Title,
			BodyType:  note.BodyType(note.BodyType_value[n.BodyType]),
			Body:      n.Body,
			CreatedAt: n.CreatedAt.Unix(),
			UpdatedAt: n.UpdatedAt.Unix(),
		}, nil
	}
}
