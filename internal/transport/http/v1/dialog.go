package v1

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	dialogListPath   = "/dialogs"
	dialogDetailPath = "/dialogs/:dialog_id"
)

type Dialog struct {
	ID          int `json:"id"`
	Participant struct {
		UserID    int    `json:"user_id"`
		Username  string `json:"username"`
		IsBlocked bool   `json:"is_blocked"`
	} `json:"participant"`
	IsBlocked bool      `json:"is_blocked"`
	CreatedAt time.Time `json:"created_at"`
}

type DialogList struct {
	Total int      `json:"total"`
	Data  []Dialog `json:"data"`
}

type DialogCreate struct {
	Participant struct {
		UserID int `json:"user_id" validate:"required"`
	} `json:"participant"`
}

type DialogUpdate struct {
	Participant struct {
		IsBlocked *bool `json:"is_blocked"`
	} `json:"participant"`
}

type DialogController struct {
}

func (dc *DialogController) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodGet, dialogListPath, dc.list)
	mux.HandlerFunc(http.MethodPost, dialogListPath, dc.create)
	mux.HandlerFunc(http.MethodGet, dialogDetailPath, dc.detail)
	mux.HandlerFunc(http.MethodPatch, dialogDetailPath, dc.update)
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
	_, _ = w, req
}

// create creates a dialog with a specified participant
//
//	@Summary	Create a dialog with a specified participant
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
	_, _ = w, req
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
	_, _ = w, req
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
	_, _ = w, req
}
