package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

type (
	role int
)

var (
	globalAuthorizer *gorbac.RBAC
)

func init() {
	globalAuthorizer = initializeRBAC()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func initializeRBAC() *gorbac.RBAC {
	rbac := gorbac.New()

	must(rbac.Add(serviceUser))
	must(rbac.Add(serviceAdmin))
	must(rbac.Add(accountAdmin))
	must(rbac.Add(accountMember))

	must(rbac.SetParent(AccountAdminRoleName, AccountMemberRoleName))
	must(rbac.SetParent(serviceAdminRoleName, AccountAdminRoleName))

	return rbac
}
