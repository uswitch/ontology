package authnz

import (
	"github.com/uswitch/ontology/pkg/middleware"
)

const UserContextKey = "authnz-user"

type Authenticator interface{
	middleware.Middleware
}
