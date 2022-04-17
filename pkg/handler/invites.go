package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nektro/mantle/pkg/db"
	"github.com/nektro/mantle/pkg/handler/controls"
	"github.com/nektro/mantle/pkg/ws"

	"github.com/nektro/go.etc/dbt"
	"github.com/nektro/go.etc/htp"
)

// InvitesMe reads info about channel
func InvitesMe(w http.ResponseWriter, r *http.Request) {
	c := htp.GetController(r)
	user := controls.GetMemberUser(c, r, w)
	usp := ws.UserPerms{}.From(user)
	if !usp.ManageInvites {
		writeAPIResponse(r, w, true, http.StatusOK, []db.Invite{})
		return
	}
	writeAPIResponse(r, w, true, http.StatusOK, db.Invite{}.All())
}

// InvitesCreate reads info about channel
func InvitesCreate(w http.ResponseWriter, r *http.Request) {
	c := htp.GetController(r)
	user := controls.GetMemberUser(c, r, w)
	usp := ws.UserPerms{}.From(user)
	c.Assert(usp.ManageInvites, "403: users require the manage_invites permission to update invites")

	nr := db.CreateInvite()
	db.CreateAudit(db.ActionInviteCreate, user, nr.UUID, "", "")
	w.WriteHeader(http.StatusCreated)
	ws.BroadcastMessage(map[string]interface{}{
		"type":   "invite-new",
		"invite": nr,
	})
}

// InviteUpdate updates info about this invite
func InviteUpdate(w http.ResponseWriter, r *http.Request) {
	c := htp.GetController(r)
	user := controls.GetMemberUser(c, r, w)
	usp := ws.UserPerms{}.From(user)
	c.Assert(usp.ManageInvites, "403: users require the manage_invites permission to update invites")

	uu := controls.GetUIDFromPath(c, r)
	iv, ok := db.QueryInviteByUID(uu)
	c.Assert(ok, "404: unable to find invite with that uuid")

	successCb := func(rs *db.Invite, pk, pv string) {
		db.CreateAudit(db.ActionInviteUpdate, user, rs.UUID, pk, pv)
		writeAPIResponse(r, w, true, http.StatusOK, map[string]interface{}{
			"invite": rs,
			"key":    pk,
			"value":  pv,
		})
		ws.BroadcastMessage(map[string]interface{}{
			"type":   "invite-update",
			"invite": rs,
			"key":    pk,
			"value":  pv,
		})
	}

	n := c.GetFormString("p_name")
	v := r.Form.Get("p_value")
	switch n {
	case "max_uses":
		_, x, err := hGrabInt(v)
		c.Assert(err == nil, "400: error parsing p_value")
		c.Assert(x >= 0, "400: p_value must be >= 0")
		c.Assert(iv.MaxUses != x, "200: property unchanged")
		iv.SetMaxUses(x)
		successCb(iv, n, v)
	case "mode":
		x, err := strconv.Atoi(v)
		c.AssertNilErr(err)
		c.Assert(iv.Mode != x, "200: property unchanged")
		iv.SetMode(x)
		successCb(iv, n, v)
	case "expires_in":
		spl := strings.SplitN(v, ",", 2)
		c.Assert(len(spl) == 2, "400: must send data in form Int,Int")
		a, err := strconv.Atoi(spl[0])
		c.AssertNilErr(err)
		b, err := strconv.Atoi(spl[1])
		c.AssertNilErr(err)
		c.Assert(iv.ExpiresIn != [...]int{a, b}, "200: property unchanged")
		iv.SetExpIn([...]int{a, b})
		successCb(iv, n, v)
	case "expires_on":
		t, err := time.Parse("2006-01-02", v)
		c.AssertNilErr(err)
		c.Assert(iv.ExpiresOn != dbt.Time(t), "200: property unchanged")
		iv.SetExpOn(t)
		successCb(iv, n, v)
	}
}

// InviteDelete updates info about this invite
func InviteDelete(w http.ResponseWriter, r *http.Request) {
	c := htp.GetController(r)
	user := controls.GetMemberUser(c, r, w)
	usp := ws.UserPerms{}.From(user)
	c.Assert(usp.ManageInvites, "403: users require the manage_invites permission to update invites")

	uu := controls.GetUIDFromPath(c, r)
	iv, ok := db.QueryInviteByUID(uu)
	c.Assert(ok, "404: unable to find invite with that uuid")

	iv.Delete()
	db.CreateAudit(db.ActionInviteDelete, user, iv.UUID, "", "")
	ws.BroadcastMessage(map[string]interface{}{
		"type":   "invite-delete",
		"invite": uu,
	})
	w.WriteHeader(http.StatusNoContent)
}
