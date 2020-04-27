package handler

import (
	"net/http"

	"github.com/nektro/mantle/pkg/db"
	"github.com/nektro/mantle/pkg/ws"

	"github.com/gorilla/mux"
)

// UsersMe is handler for /api/users/@me
func UsersMe(w http.ResponseWriter, r *http.Request) {
	_, user, err := apiBootstrapRequireLogin(r, w, http.MethodGet, true)
	if err != nil {
		return
	}
	writeAPIResponse(r, w, true, http.StatusOK, map[string]interface{}{
		"me":    user,
		"perms": ws.UserPerms{}.From(user),
	})
}

// UsersRead is handler for /api/users/{uuid}
func UsersRead(w http.ResponseWriter, r *http.Request) {
	_, _, err := apiBootstrapRequireLogin(r, w, http.MethodGet, true)
	if err != nil {
		return
	}
	uu := mux.Vars(r)["uuid"]
	u, ok := db.QueryUserByUUID(uu)
	writeAPIResponse(r, w, ok, http.StatusOK, u)
}

// UsersOnline is handler for /api/users/online
func UsersOnline(w http.ResponseWriter, r *http.Request) {
	_, _, err := apiBootstrapRequireLogin(r, w, http.MethodGet, true)
	if err != nil {
		return
	}
	writeAPIResponse(r, w, true, http.StatusOK, ws.AllOnlineIDs())
}

// UserUpdate is handler for /api/users/{uuid}/update
func UserUpdate(w http.ResponseWriter, r *http.Request) {
	_, user, err := apiBootstrapRequireLogin(r, w, http.MethodPut, true)
	if err != nil {
		return
	}
	uu := mux.Vars(r)["uuid"]
	u, ok := db.QueryUserByUUID(uu)
	if !ok {
		return
	}
	if hGrabFormStrings(r, w, "p_name") != nil {
		return
	}

	successCb := func(us *db.User, pk, pv string) {
		db.CreateAudit(db.ActionUserUpdate, user, us.UUID, pk, pv)
		writeAPIResponse(r, w, true, http.StatusOK, map[string]interface{}{
			"user":  us,
			"key":   pk,
			"value": pv,
		})
		ws.BroadcastMessage(map[string]interface{}{
			"type":  "user-update",
			"user":  us,
			"key":   pk,
			"value": pv,
		})
	}

	n := r.Form.Get("p_name")
	v := r.Form.Get("p_value")
	up := ws.UserPerms{}.From(user)
	if n != "nickname" {
		if len(v) == 0 {
			writeAPIResponse(r, w, false, http.StatusBadRequest, "missing form value 'p_value'.")
			return
		}
	}
	switch n {
	case "nickname":
		if user.UUID != u.UUID {
			return
		}
		u.SetNickname(v)
		successCb(u, n, v)
	case "add_role":
		rl, ok := db.QueryRoleByUID(v)
		if !ok {
			return
		}
		if !up.ManageRoles {
			return
		}
		if user.GetRolesSorted()[0].Position >= rl.Position {
			return
		}
		u.AddRole(v)
		successCb(u, n, v)
	case "remove_role":
		rl, ok := db.QueryRoleByUID(v)
		if !ok {
			return
		}
		if !up.ManageRoles {
			return
		}
		if user.GetRolesSorted()[0].Position >= rl.Position {
			return
		}
		u.RemoveRole(v)
		successCb(u, n, v)
	}
}
