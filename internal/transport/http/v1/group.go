package v1

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	groupListPath   = "/groups"
	groupDetailPath = "/groups/:group_id"
)

type Group struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type GroupList struct {
	Total int     `json:"total"`
	Data  []Group `json:"data"`
}

type GroupUpdate struct {
	Name        string `json:"name"        validate:"required,max=255"`
	Description string `json:"description" validate:"max=10000"`
}

type GroupCreate struct {
	GroupUpdate
}

type GroupController struct {
}

func (gc *GroupController) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodGet, groupListPath, gc.list)
	mux.HandlerFunc(http.MethodPost, groupListPath, gc.create)
	mux.HandlerFunc(http.MethodGet, groupDetailPath, gc.detail)
	mux.HandlerFunc(http.MethodPut, groupDetailPath, gc.update)
	mux.HandlerFunc(http.MethodDelete, groupDetailPath, gc.delete)
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
	_, _ = w, req
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
	_, _ = w, req
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
	_, _ = w, req
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
	_, _ = w, req
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
	_, _ = w, req
}
