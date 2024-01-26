package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/httputil"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/julienschmidt/httprouter"
)

const (
	groupParticipantListPath   = "/participants"
	groupParticipantDetailPath = "/participants/:user_id"
)

type GroupParticipant struct {
	UserID  int                           `json:"user_id"`
	Status  entity.GroupParticipantStatus `json:"status"`
	IsAdmin bool                          `json:"is_admin"`
}

func NewGroupParticipant(participant entity.GroupParticipant) GroupParticipant {
	return GroupParticipant{
		UserID:  participant.UserID,
		Status:  participant.Status,
		IsAdmin: participant.IsAdmin,
	}
}

type GroupParticipantList struct {
	Total int                `json:"total"`
	Data  []GroupParticipant `json:"data"`
}

func NewGroupParticipantList(participants []entity.GroupParticipant) GroupParticipantList {
	data := make([]GroupParticipant, len(participants))
	for i, participant := range participants {
		data[i] = NewGroupParticipant(participant)
	}

	return GroupParticipantList{
		Total: len(participants),
		Data:  data,
	}
}

type GroupParticipantUpdate struct {
	Status *string `json:"status" validate:"omitempty,oneof=joined left kicked"`
}

//go:generate mockery --inpackage --testonly --case underscore --name GroupParticipantService
type GroupParticipantService interface {
	List(ctx context.Context, groupID int) ([]entity.GroupParticipant, error)
	Get(ctx context.Context, groupID, userID int) (entity.GroupParticipant, error)
	Invite(ctx context.Context, groupID, userID int) (entity.GroupParticipant, error)
	UpdateStatus(ctx context.Context, groupID, userID int, status entity.GroupParticipantStatus) error
}

type GroupParticipantControllerConfig struct {
	Service   GroupParticipantService
	Authorize middleware.Middleware
	Validator validator.Validator
}

type GroupParticipantController struct {
	service   GroupParticipantService
	authorize middleware.Middleware
	validator validator.Validator
}

func NewGroupParticipantController(conf GroupParticipantControllerConfig) *GroupParticipantController {
	return &GroupParticipantController{
		service:   conf.Service,
		authorize: conf.Authorize,
		validator: conf.Validator,
	}
}

func (pc *GroupParticipantController) Register(mux *httprouter.Router) {
	mux.Handler(http.MethodGet, groupDetailPath+groupParticipantListPath, pc.authorize(http.HandlerFunc(pc.list)))
	mux.Handler(http.MethodPost, groupDetailPath+groupParticipantListPath, pc.authorize(http.HandlerFunc(pc.invite)))
	mux.Handler(http.MethodGet, groupDetailPath+groupParticipantDetailPath, pc.authorize(http.HandlerFunc(pc.detail)))
	mux.Handler(http.MethodPatch, groupDetailPath+groupParticipantDetailPath, pc.authorize(http.HandlerFunc(pc.update)))
}

// list lists all participants for a specified group
//
//	@Summary	List all participants for a specified group
//	@Tags		group-participants
//	@Accept		json
//	@Produce	json
//	@Param		group_id	path		int	true	"Group identity"
//	@Success	200			{object}	GroupParticipantList
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/groups/{group_id}/participants  [get]
func (pc *GroupParticipantController) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	groupID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(groupIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	participants, err := pc.service.List(ctx, groupID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewGroupParticipantList(participants))
}

// invite invites a specified participant in a group
//
//	@Summary	Invite a specified participant in a group
//	@Tags		group-participants
//	@Accept		json
//	@Produce	json
//	@Param		group_id	path		int	true	"Group identity"
//	@Param		user_id		query		int	true	"User identity to invite"
//	@Success	201			{object}	GroupParticipant
//	@Failure	400			{object}	httputil.Error
//	@Failure	403			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/groups/{group_id}/participants  [post]
func (pc *GroupParticipantController) invite(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	groupID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(groupIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}
	userID, err := strconv.Atoi(req.URL.Query().Get(userIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodeQueryParamsFailed.Wrap(err))
		return
	}

	participant, err := pc.service.Invite(ctx, groupID, userID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		case errors.Is(err, entity.ErrAddNonExistentUserToGroup):
			httputil.RespondError(ctx, w, errInviteNonExistentUserToGroup.Wrap(err))
		case errors.Is(err, entity.ErrSuchGroupParticipantAlreadyExists):
			httputil.RespondError(ctx, w, errSuchGroupParticipantAlreadyExists.Wrap(err))
		case errors.Is(err, entity.ErrForbiddenPerformAction):
			httputil.RespondError(ctx, w, httputil.ErrForbiddenPerformAction.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusCreated, NewGroupParticipant(participant))
}

// detail gets a specified participant in a group
//
//	@Summary	Get a specified participant in a group
//	@Tags		group-participants
//	@Accept		json
//	@Produce	json
//	@Param		group_id	path		int	true	"Group identity"
//	@Param		user_id		path		int	true	"User identity"
//	@Success	200			{object}	GroupParticipant
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/groups/{group_id}/participants/{user_id}  [get]
func (pc *GroupParticipantController) detail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	groupID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(groupIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}
	userID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(userIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	participant, err := pc.service.Get(ctx, groupID, userID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		case errors.Is(err, entity.ErrGroupParticipantNotFound):
			httputil.RespondError(ctx, w, errGroupParticipantNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewGroupParticipant(participant))
}

// update updates a specified participant in a group
//
//	@Summary		Update a specified participant in a group
//	@Description	It can be used to join/kick/leave participant from the group.
//	@Tags			group-participants
//	@Accept			json
//	@Produce		json
//	@Param			group_id	path	int						true	"Group identity"
//	@Param			user_id		path	int						true	"User identity"
//	@Param			input		body	GroupParticipantUpdate	true	"Body to update"
//	@Success		204			"No Content"
//	@Failure		400			{object}	httputil.Error
//	@Failure		403			{object}	httputil.Error
//	@Failure		404			{object}	httputil.Error
//	@Failure		500			{object}	httputil.Error
//	@Security		JWTAuth
//	@Router			/groups/{group_id}/participants/{user_id}  [patch]
func (pc *GroupParticipantController) update(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	groupID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(groupIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}
	userID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(userIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	var bodyObj GroupParticipantUpdate

	if err = httputil.DecodeBody(req.Body, &bodyObj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err = pc.validator.Struct(bodyObj); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	if bodyObj.Status == nil {
		httputil.RespondSuccess(ctx, w, http.StatusNoContent, nil)
		return
	}

	err = pc.service.UpdateStatus(ctx, groupID, userID, entity.GroupParticipantStatus(*bodyObj.Status))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		case errors.Is(err, entity.ErrGroupParticipantNotFound):
			httputil.RespondError(ctx, w, errGroupParticipantNotFound.Wrap(err))
		case errors.Is(err, entity.ErrIncorrectGroupParticipantStatusTransit):
			httputil.RespondError(ctx, w, errIncorrectGroupParticipantStatusTransit)
		case errors.Is(err, entity.ErrForbiddenPerformAction):
			httputil.RespondError(ctx, w, httputil.ErrForbiddenPerformAction.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusNoContent, nil)
}
