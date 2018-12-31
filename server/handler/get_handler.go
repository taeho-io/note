package handler

import (
	"database/sql"

	"github.com/taeho-io/auth"
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

func Get(db *sql.DB, authCli auth.AuthClient) GetHandlerFunc {
	return func(ctx context.Context, req *note.GetRequest) (*note.GetResponse, error) {
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

		claims, _ := token.GetClaimsFromMD(ctx)
		if claims != nil && claims.UserID != n.CreatedBy {
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
