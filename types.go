package mrequest

type userAgentType string

const (
	MobileDevice  userAgentType = "mobile"
	DesktopDevice userAgentType = "desktop"
	IOTDevice     userAgentType = "iot"
	All           userAgentType = "all"
)

//------------------------
type contentType string

const (
	JSON     contentType = "application/json"
	TextHTML contentType = "text/html"
	FormType contentType = "application/x-www-form-urlencoded"
)
