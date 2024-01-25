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
	dialogListPath   = "/api/v1/dialogs"
	dialogDetailPath = "/api/v1/dialogs/:dialog_id"
)

const (
	dialogIDPathParam = "dialog_id"
)

type DialogPartner struct {
	UserID    int  `json:"user_id"`
	IsBlocked bool `json:"is_blocked"`
}

type Dialog struct {
	ID        int           `json:"id"`
	IsBlocked bool          `json:"is_blocked"`
	Partner   DialogPartner `json:"partner"`
	CreatedAt time.Time     `json:"created_at"`
}

func NewDialog(dialog entity.Dialog) Dialog {
	return Dialog{
		ID:        dialog.ID,
		IsBlocked: dialog.IsBlocked,
		Partner: DialogPartner{
			UserID:    dialog.Partner.UserID,
			IsBlocked: dialog.Partner.IsBlocked,
		},
		CreatedAt: dialog.CreatedAt,
	}
}

type DialogList struct {
	Total int      `json:"total"`
	Data  []Dialog `json:"data"`
}

func NewDialogList(dialogs []entity.Dialog) DialogList {
	data := make([]Dialog, len(dialogs))
	for i, dialog := range dialogs {
		data[i] = NewDialog(dialog)
	}

	return DialogList{
		Total: len(dialogs),
		Data:  data,
	}
}

type DialogCreate struct {
	Partner struct {
		UserID int `json:"user_id" validate:"required"`
	} `json:"partner"`
}

func (d DialogCreate) DTO() dto.DialogCreate {
	return dto.DialogCreate{PartnerUserID: d.Partner.UserID}
}

type DialogUpdate struct {
	Partner struct {
		IsBlocked *bool `json:"is_blocked"`
	} `json:"partner"`
}

func (d DialogUpdate) DTO() dto.DialogUpdate {
	return dto.DialogUpdate{
		PartnerIsBlocked: d.Partner.IsBlocked,
	}
}

//go:generate mockery --inpackage --testonly --case underscore --name DialogService
type DialogService interface {
	List(ctx context.Context) ([]entity.Dialog, error)
	Create(ctx context.Context, obj dto.DialogCreate) (entity.Dialog, error)
	GetByID(ctx context.Context, id int) (entity.Dialog, error)
	Update(ctx context.Context, obj dto.DialogUpdate) error
}

type DialogControllerConfig struct {
	Service   DialogService
	Authorize middleware.Middleware
	Validator validator.Validator
}

type DialogController struct {
	service   DialogService
	authorize middleware.Middleware
	validator validator.Validator
}

func NewDialogController(conf DialogControllerConfig) *DialogController {
	return &DialogController{
		service:   conf.Service,
		authorize: conf.Authorize,
		validator: conf.Validator,
	}
}

func (dc *DialogController) Register(mux *httprouter.Router) {
	mux.Handler(http.MethodGet, dialogListPath, dc.authorize(http.HandlerFunc(dc.list)))
	mux.Handler(http.MethodPost, dialogListPath, dc.authorize(http.HandlerFunc(dc.create)))
	mux.Handler(http.MethodGet, dialogDetailPath, dc.authorize(http.HandlerFunc(dc.detail)))
	mux.Handler(http.MethodPatch, dialogDetailPath, dc.authorize(http.HandlerFunc(dc.update)))
}

// list lists all dialogs
//
//	@Summary	List all dialogs
//	@Tags		dialogs
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	DialogList
//	@Failure	500	{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/dialogs  [get]
func (dc *DialogController) list(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	dialogs, err := dc.service.List(ctx)
	if err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewDialogList(dialogs))
}

// create creates a dialog with a specified partner
//
//	@Summary	Create a dialog with a specified partner
//	@Tags		dialogs
//	@Accept		json
//	@Produce	json
//	@Param		input	body		DialogCreate	true	"Body to create"
//	@Success	201		{object}	Dialog
//	@Failure	400		{object}	httputil.Error
//	@Failure	500		{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/dialogs  [post]
func (dc *DialogController) create(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var bodyObj DialogCreate

	if err := httputil.DecodeBody(req.Body, &bodyObj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	if err := dc.validator.Struct(bodyObj); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			httputil.RespondError(ctx, w, httputil.ErrValidationFailed.WithData(ve.Fields).Wrap(err))
			return
		}

		httputil.RespondError(ctx, w, err)
		return
	}

	dialog, err := dc.service.Create(ctx, bodyObj.DTO())
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrSuchDialogAlreadyExists):
			httputil.RespondError(ctx, w, errSuchDialogAlreadyExists.Wrap(err))
		case errors.Is(err, entity.ErrCreateDialogWithYourself):
			httputil.RespondError(ctx, w, errCreateDialogWithYourself.Wrap(err))
		case errors.Is(err, entity.ErrCreateDialogWithNonExistentUser):
			httputil.RespondError(ctx, w, errCreateDialogWithNonExistenceUser.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusCreated, NewDialog(dialog))
}

// detail gets a specified dialog
//
//	@Summary	Get a specified dialog
//	@Tags		dialogs
//	@Accept		json
//	@Produce	json
//	@Param		dialog_id	path		int	true	"Dialog identity"
//	@Success	200			{object}	Dialog
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/dialogs/{dialog_id}  [get]
func (dc *DialogController) detail(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	dialogID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(dialogIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	dialog, err := dc.service.GetByID(ctx, dialogID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrDialogNotFound):
			httputil.RespondError(ctx, w, errDialogNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusOK, NewDialog(dialog))
}

// update updates a specified dialog
//
//	@Summary	Update a specified dialog
//	@Tags		dialogs
//	@Accept		json
//	@Produce	json
//	@Param		dialog_id	path	int				true	"Dialog identity"
//	@Param		input		body	DialogUpdate	true	"Body to update"
//	@Success	204			"No Content"
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Security	JWTAuth
//	@Router		/dialogs/{dialog_id}  [patch]
func (dc *DialogController) update(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	dialogID, err := strconv.Atoi(httprouter.ParamsFromContext(ctx).ByName(dialogIDPathParam))
	if err != nil {
		httputil.RespondError(ctx, w, httputil.ErrDecodePathParamsFailed.Wrap(err))
		return
	}

	var bodyObj DialogUpdate

	if err = httputil.DecodeBody(req.Body, &bodyObj); err != nil {
		httputil.RespondError(ctx, w, err)
		return
	}

	obj := bodyObj.DTO()
	obj.ID = dialogID

	err = dc.service.Update(ctx, obj)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrDialogNotFound):
			httputil.RespondError(ctx, w, errDialogNotFound.Wrap(err))
		default:
			httputil.RespondError(ctx, w, err)
		}

		return
	}

	httputil.RespondSuccess(ctx, w, http.StatusNoContent, nil)
}
