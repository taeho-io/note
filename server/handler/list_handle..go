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

type ListHandlerFunc func(ctx context.Context, request *note.ListRequest) (*note.ListResponse, error)

func noteMessageFromModel(n *models.Note) *note.NoteMessage {
	return &note.NoteMessage{
		NoteId:    n.ID,
		CreatedBy: n.CreatedBy,
		Title:     n.Title,
		BodyType:  note.BodyType(note.BodyType_value[n.BodyType]),
		Body:      n.Body,
		CreatedAt: n.CreatedAt.Unix(),
		UpdatedAt: n.UpdatedAt.Unix(),
	}
}

func List(db *sql.DB) ListHandlerFunc {
	return func(ctx context.Context, req *note.ListRequest) (*note.ListResponse, error) {
		claims, err := token.GetClaimsFromMD(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		userID := claims.UserID

		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		notes, err := models.Notes(
			Where("created_by=?", userID),
			Offset(int(req.Offset)),
			Limit(int(req.Limit)),
			OrderBy("updated_at DESC"),
		).All(ctx, db)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		var noteMessages []*note.NoteMessage
		for _, n := range notes {
			noteMessages = append(noteMessages, noteMessageFromModel(n))
		}

		return &note.ListResponse{
			Notes: noteMessages,
		}, nil
	}
}
