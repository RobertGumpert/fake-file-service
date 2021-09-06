package streaming

import "errors"

type (
	Context struct {
		ClientRemoteAddress RemoteAddress
		Message             []byte
		Error               error
	}
	URI           string
	RemoteAddress string
	Handler       func(context *Context)
)

var (
	ErrorPoolClientIsFilled = errors.New("Error: pool client is filled")
	ErrorClientObjectIsNil  = errors.New("Error: client object is nil")
	ErrorHandlerIsntExist   = errors.New("Error: handler isn't exist")
)
