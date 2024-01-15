package sessions

import (
	"errors"
	"fmt"

	fiberSession "github.com/gofiber/fiber/v2/middleware/session"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

var (
	errAnonymousSession = errors.New("session is anonymous")
)

// Session define a read only authenticated user session.
type Session interface {
	UserId() users.UserId
}

func userSessionFromSession(fiberSession *fiberSession.Session) (userSession, error) {
	uid := fiberSession.Get(UserIdKey)
	if uid == nil {
		return userSession{}, fmt.Errorf("%w: no user id in session", errAnonymousSession)
	}

	uidStr, isString := uid.(string)
	if !isString {
		return userSession{}, fmt.Errorf("%w: invalid user id, not a string", errAnonymousSession)
	}

	userId, err := users.ParseUserId(uidStr)
	if err != nil {
		return userSession{}, fmt.Errorf("%w: failed to parse user id: %w", errAnonymousSession, err)
	}

	return userSession{
		userId,
		fiberSession,
	}, nil
}

type userSession struct {
	userId users.UserId
	*fiberSession.Session
}

// UserId returns associated user id.
func (us userSession) UserId() users.UserId {
	return us.userId
}
