package v1

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	groupParticipantListPath   = "/participants"
	groupParticipantDetailPath = "/participants/:user_id"
)

type GroupParticipantDetail struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Status   string `json:"status"`
	IsAdmin  bool   `json:"is_admin"`
}

type GroupParticipantList struct {
	Total int                      `json:"total"`
	Data  []GroupParticipantDetail `json:"data"`
}

type GroupParticipantUpdate struct {
	Status *string `json:"status" validate:"oneof=joined left kicked"`
}

type GroupParticipantController struct {
}

func (pc *GroupParticipantController) Register(mux *httprouter.Router) {
	mux.HandlerFunc(http.MethodGet, groupDetailPath+groupParticipantListPath, pc.list)
	mux.HandlerFunc(http.MethodGet, groupDetailPath+groupParticipantDetailPath, pc.detail)
	mux.HandlerFunc(http.MethodGet, groupDetailPath+groupParticipantDetailPath, pc.update)
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
//	@Router		/groups/{group_id}/participants  [get]
func (pc *GroupParticipantController) list(w http.ResponseWriter, req *http.Request) {

}

// detail gets a specified participant in a group
//
//	@Summary	Get a specified participant in a group
//	@Tags		group-participants
//	@Accept		json
//	@Produce	json
//	@Param		group_id	path		int	true	"Group identity"
//	@Param		user_id		path		int	true	"User identity"
//	@Success	200			{object}	GroupParticipantDetail
//	@Failure	400			{object}	httputil.Error
//	@Failure	404			{object}	httputil.Error
//	@Failure	500			{object}	httputil.Error
//	@Router		/groups/{group_id}/participants/{user_id}  [get]
func (pc *GroupParticipantController) detail(w http.ResponseWriter, req *http.Request) {

}

// update updates a specified participant in a group
//
//	@Summary		Update a specified participant in a group
//	@Description	It can be used to join/kick/leave participant from the group.
//	@Tags			group-participants
//	@Accept			json
//	@Produce		json
//	@Param			group_id	path	int	true	"Group identity"
//	@Param			user_id		path	int	true	"User identity"
//	@Success		204			"No Content"
//	@Failure		400			{object}	httputil.Error
//	@Failure		403			{object}	httputil.Error
//	@Failure		404			{object}	httputil.Error
//	@Failure		500			{object}	httputil.Error
//	@Router			/groups/{group_id}/participants/{user_id}  [patch]
func (pc *GroupParticipantController) update(w http.ResponseWriter, req *http.Request) {

}
