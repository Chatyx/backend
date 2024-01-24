package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/httputil"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/julienschmidt/httprouter"
)

const (
	groupListPath   = "/api/v1/groups"
	groupDetailPath = "/api/v1/groups/:group_id"
)

const (
	groupIDPathParam = "group_id"
)

type Group struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewGroup(group entity.Group) Group {
	return Group{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
	}
}

type GroupList struct {
	Total int     `json:"total"`
	Data  []Group `json:"data"`
}

func NewGroupList(groups []entity.Group) GroupList {
	data := make([]Group, len(groups))
	for i, group := range groups {
		data[i] = NewGroup(group)
	}

	return GroupList{
		Total: len(groups),
		Data:  data,
	}
}

type GroupCreate struct {
	Name        string `json:"name"        validate:"required,max=255"`
	Description string `json:"description" validate:"max=10000"`
}

func (g GroupCreate) DTO() dto.GroupCreate {
	return dto.GroupCreate{
		Name:        g.Name,
		Description: g.Description,
	}
}

type GroupUpdate struct {
	Name        string `json:"name"        validate:"required,max=255"`
	Description string `json:"description" validate:"max=10000"`
}

func (g GroupUpdate) DTO() dto.GroupUpdate {
	return dto.GroupUpdate{
		Name:        g.Name,
		Description: g.Description,
	}
}

//go:generate mockery --inpackage --testonly --case underscore --name GroupService
type GroupService interface {
	List(ctx context.Context) ([]entity.Group, error)
	Create(ctx context.Context, obj dto.GroupCreate) (entity.Group, error)
	GetByID(ctx context.Context, id int) (entity.Group, error)
	Update(ctx context.Context, obj dto.GroupUpdate) (entity.Group, error)
	Delete(ctx context.Context, id int) error
}

type GroupControllerConfig struct {
	Service   GroupService
	Authorize middleware.Middleware
	Validator validator.Validator
}

type GroupController struct {
	service   GroupService
	authorize middleware.Middleware
	validator validator.Validator
}

func NewGroupController(conf GroupControllerConfig) *GroupController {
	return &GroupController{
		service:   conf.Service,
		authorize: conf.Authorize,
		validator: conf.Validator,
	}
}

func (gc *GroupController) Register(mux *httprouter.Router) {
	mux.Handler(http.MethodGet, groupListPath, gc.authorize(http.HandlerFunc(gc.list)))
	mux.Handler(http.MethodPost, groupListPath, gc.authorize(http.HandlerFunc(gc.create)))
	mux.Handler(http.MethodGet, groupDetailPath, gc.authorize(http.HandlerFunc(gc.detail)))
	mux.Handler(http.MethodPut, groupDetailPath, gc.authorize(http.HandlerFunc(gc.update)))
	mux.Handler(http.MethodDelete, groupDetailPath, gc.authorize(http.HandlerFunc(gc.delete)))
}

// list lists all groups
//
//	@Summary	List all groups
//	@Tags		groups
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	GroupList
//	@Failure	500	{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/groups  [get]
func (gc *GroupController) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	groups, err := gc.service.List(ctx)
	if err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewGroupList(groups))
}

// create creates a group
//
//	@Summary	Create a group
//	@Tags		groups
//	@Accept		json
//	@Produce	json
//	@Param		input	body		GroupCreate	true	"Body to create"
//	@Success	201		{object}	Group
//	@Failure	400		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/groups  [post]
func (gc *GroupController) create(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var bodyObj GroupCreate

	if err := httputil.DecodeBody(req.Body, &bodyObj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err := gc.validator.Struct(bodyObj); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	group, err := gc.service.Create(ctx, bodyObj.DTO())
	if err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusCreated, NewGroup(group))
}

// detail gets a specified group
//
//	@Summary	Get a specified group
//	@Tags		groups
//	@Accept		json
//	@Produce	json
//	@Param		group_id	path		int	true	"Group identity"
//	@Success	200			{object}	Group
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/groups/{group_id}  [get]
func (gc *GroupController) detail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	groupID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(groupIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	group, err := gc.service.GetByID(ctx, groupID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewGroup(group))
}

// update updates a specified group
//
//	@Summary	Update a specified group
//	@Tags		groups
//	@Accept		json
//	@Produce	json
//	@Param		group_id	path		int			true	"Group identity"
//	@Param		input		body		GroupUpdate	true	"Body to update"
//	@Success	200			{object}	Group
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/groups/{group_id}  [put]
func (gc *GroupController) update(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	groupID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(groupIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	var bodyObj GroupUpdate

	if err = httputil.DecodeBody(req.Body, &bodyObj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err = gc.validator.Struct(bodyObj); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	obj := bodyObj.DTO()
	obj.ID = groupID

	group, err := gc.service.Update(ctx, obj)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewGroup(group))
}

// delete deletes a specified group
//
//	@Summary	Delete a specified group
//	@Tags		groups
//	@Accept		json
//	@Produce	json
//	@Param		group_id	path	int	true	"Group identity"
//	@Success	204			"No Content"
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/groups/{group_id}  [delete]
func (gc *GroupController) delete(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	groupID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(groupIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	if err = gc.service.Delete(ctx, groupID); err != nil {
		switch {
		case errors.Is(err, entity.ErrGroupNotFound):
			httputil.RespondError(ctx, w, errGroupNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusNoContent, nil)
}
