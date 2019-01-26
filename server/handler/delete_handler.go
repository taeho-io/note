package handler

import (
	"database/sql"

	"github.com/taeho-io/auth/pkg/token"
	"github.com/taeho-io/note"
	"github.com/taeho-io/note/server/models"
	. "github.com/volatiletech/sqlboiler/queries/qm"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteHandlerFunc func(ctx context.Context, request *note.DeleteRequest) (*note.DeleteResponse, error)

func Delete(db *sql.DB) DeleteHandlerFunc {
	return func(ctx context.Context, req *note.DeleteRequest) (*note.DeleteResponse, error) {
		claims, err := token.GetClaimsFromMD(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		userID := claims.UserID

		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if _, err := models.Notes(
			Where("id=?", req.NoteId),
			And("created_by=?", userID),
		).DeleteAll(ctx, db); err != nil {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		return &note.DeleteResponse{}, nil
	}
}
