package event

const AUTHORIZATION_REGISTER = "authorization.register"
const AUTHORIZATION_LOGIN = "authorization.login"
const AUTHORIZATION_LOGOUT = "authorization.logout"
const AUTHORIZATION_REFRESH_TOKEN = "authorization.refresh-token"
const AUTHORIZATION_VALIDATE_TOKEN = "authorization.validate-token"
const AUTHORIZATION_REQUEST_PASSWORD_RESET = "authorization.request-password-reset"

type AuthorizationRegisterMessage struct {
	ID string
}

type AuthorizationLoginMessage struct {
	ID string
}

type AuthorizationLogoutMessage struct {
	ID string
}

type AuthorizationRefreshTokenMessage struct {
	ID string
}

type AuthorizationValidateTokenMessage struct {
	ID string
}

type AuthorizationRequestPasswordReset struct {
	ID string
}
