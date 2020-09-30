package db

import (
	"database/sql"
	"sort"
	"strconv"

	"github.com/nektro/go-util/arrays/stringsu"
	"github.com/nektro/go-util/util"
	dbstorage "github.com/nektro/go.dbstorage"
	"github.com/nektro/go.etc/store"

	. "github.com/nektro/go.etc/dbt"
)

type User struct {
	ID         int64  `json:"id"`
	Provider   string `json:"provider" dbsorm:"1"`
	Snowflake  string `json:"snowflake" dbsorm:"1"`
	UUID       UUID   `json:"uuid" dbsorm:"1"`
	IsMember   bool   `json:"is_member" dbsorm:"1"`
	IsBanned   bool   `json:"is_banned" dbsorm:"1"`
	Name       string `json:"name" dbsorm:"1"`
	Nickname   string `json:"nickname" dbsorm:"1"`
	JoindedOn  Time   `json:"joined_on" dbsorm:"1"`
	LastActive Time   `json:"last_active" dbsorm:"1"`
	Roles      List   `json:"roles" dbsorm:"1"`
}

//
//

func QueryUserByUUID(uid UUID) (*User, bool) {
	ch, ok := dbstorage.ScanFirst(db.Build().Se("*").Fr(cTableUsers).Wh("uuid", uid.String()), User{}).(*User)
	return ch, ok
}

func QueryUserBySnowflake(provider string, flake string, name string) *User {
	us, ok := dbstorage.ScanFirst(db.Build().Se("*").Fr(cTableUsers).Wh("provider", provider).Wh("snowflake", flake), User{}).(*User)
	if ok {
		return us
	}
	store.This.Lock()
	defer store.This.Unlock()
	//
	id := db.QueryNextID(cTableUsers)
	uid := NewUUID()
	co := now()
	roles := List{}
	if id == 1 {
		roles = append(roles, "o")
		Props.Set("owner", uid.String())
	}
	u := &User{id, provider, flake, uid, false, false, name, "", co, co, roles}
	db.Build().InsI(cTableUsers, u).Exe()
	Props.Increment("count_" + cTableUsers)
	return u
}

//
//

// Scan implements dbstorage.Scannable
func (v User) Scan(rows *sql.Rows) dbstorage.Scannable {
	rows.Scan(&v.ID, &v.Provider, &v.Snowflake, &v.UUID, &v.IsMember, &v.IsBanned, &v.Name, &v.Nickname, &v.JoindedOn, &v.LastActive, &v.Roles)
	return &v
}

// Count returns the total number of Users
func (v User) Count() int64 {
	return db.QueryRowCount(cTableUsers)
}

func (v User) MemberCount() int64 {
	rows := db.Build().Se("count(*)").Fr(cTableUsers).Wh("is_member", "1").Exe()
	defer rows.Close()
	c := int64(0)
	rows.Next()
	rows.Scan(&c)
	return c
}

//
//

func (v *User) i() string {
	return v.UUID.String()
}

func (v User) t() string {
	return cTableUsers
}

func (v User) b() dbstorage.QueryBuilder {
	return db.Build().Se("*").Fr(v.t())
}


//
// searchers
//

//
// modifiers
//

func (u *User) SetAsMember(b bool) {
	m := u.IsMember
	doUp(u, "is_member", strconv.Itoa(util.Btoi(b)))
	if !m && b {
		Props.Increment("count_users_members")
	}
	if m && !b {
		Props.Decrement("count_users_members")
	}
	u.IsMember = b
}

func (u *User) SetName(s string) {
	doUp(u, "name", s)
	u.Name = s
}

// DeleteMessage attempts to delete a UID from this Channel's associated message
// table. If the UID is not a message in this Channel, nothing happens.
func (u *User) DeleteMessage(c *Channel, uid UUID) {
	tn := cTableMessagesPrefix + c.i()
	r := db.Build().Se("count(*)").Fr(tn).Wh("uuid", uid.String()).Wh("author", u.UUID.String()).Exe()
	l := 0
	r.Scan(&l)
	r.Close()
	Props.Decrement("count_" + tn)
	Props.Increment("count_" + tn + "_deletes")
	db.Build().Del(tn).Wh("uuid", uid.String()).Wh("author", u.UUID.String()).Exe()
}

func (u *User) HasRole(role UUID) bool {
	return stringsu.Contains(u.Roles, role.String())
}

func (u *User) AddRole(role UUID) {
	if u.HasRole(role) {
		return
	}
	u.Roles = append(u.Roles, role.String())
	doUp(u, "roles", u.Roles.String())
}

func (u *User) RemoveRole(role UUID) {
	if !u.HasRole(role) {
		return
	}
	u.Roles = stringsu.Remove(u.Roles, role.String())
	doUp(u, "roles", u.Roles.String())
}

func (u *User) GetRoles() []*Role {
	res := []*Role{}
	for _, item := range u.Roles {
		r, ok := QueryRoleByUID(UUID(item))
		if !ok {
			continue
		}
		res = append(res, r)
	}
	return res
}

func (u *User) GetRolesSorted() []*Role {
	res := u.GetRoles()
	sort.Slice(res, func(i, j int) bool {
		return res[i].Position < res[j].Position
	})
	return res
}

func (u *User) SetUID(uid UUID) {
	oid := u.UUID
	db.Build().Up(cTableUsers, "uuid", uid.String()).Wh("uuid", u.UUID.String()).Exe()
	for _, item := range (Channel{}.All()) {
		db.Build().Up(cTableMessagesPrefix+item.UUID.String(), "author", uid.String()).Wh("author", u.UUID.String()).Exe()
	}
	u.UUID = uid
	util.Log("user-update:", "updated", u.Name+"#"+strconv.FormatInt(u.ID, 10), "from", oid, "to", u.UUID)
}

func (u *User) ResetUID() {
	u.SetUID(NewUUID())
	if u.HasRole("o") {
		Props.Set("owner", u.UUID.String())
	}
}

func (u *User) SetNickname(s string) {
	doUp(u, "nickname", s)
	u.Nickname = s
}
