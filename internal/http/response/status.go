package response

const (
	StatusSuccess             = "success"
	StatusValidationError     = "validation_error"
	StatusUnauthorizedAccess  = "unauthorized_access"
	StatusResourceNotFound    = "resource_not_found"
	StatusUnprocessableEntity = "unprocessable_entity"
	StatusSessionExpired      = "session_expired"
	StatusInternalServerError = "internal_server_error"
)

func StatusFromCode(code int) string {
	switch code {
	case 200:
		return StatusSuccess
	case 400:
		return StatusValidationError
	case 401:
		return StatusUnauthorizedAccess
	case 404:
		return StatusResourceNotFound
	case 419:
		return StatusSessionExpired
	case 422:
		return StatusUnprocessableEntity
	default:
		return StatusInternalServerError
	}
}
