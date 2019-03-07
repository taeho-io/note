package handler

import (
	"database/sql"

	"github.com/taeho-io/auth/pkg/token"
	"github.com/taeho-io/idl/gen/go/note"
	"github.com/taeho-io/note/server/models"
	"github.com/volatiletech/sqlboiler/boil"
	. "github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateHandlerFunc func(ctx context.Context, req *note.UpdateRequest) (*note.UpdateResponse, error)

func Update(db *sql.DB) UpdateHandlerFunc {
	return func(ctx context.Context, req *note.UpdateRequest) (*note.UpdateResponse, error) {
		claims, err := token.GetClaimsFromMD(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		userID := claims.UserID

		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		n, err := models.Notes(
			Where("id=?", req.NoteId),
			And("created_by=?", userID),
		).One(ctx, db)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return nil, ErrNoteNotFound
			}
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		n.Title = req.Title
		n.BodyType = req.BodyType.String()
		n.Body = req.Body
		if _, err := n.Update(ctx, db, boil.Infer()); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &note.UpdateResponse{}, nil
	}
}
