package user

import (
	"errors"
	"net/http"

	hspb "github.com/anousonefs/golang-htmx-template/internal/proto/http"

	"google.golang.org/genproto/googleapis/rpc/code"
	edpb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUnauthorized         = errors.New("unauthorized")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrNoInfo               = errors.New("no info")
	ErrUnProcessAbleEntity  = errors.New("unprocessable entity")
	ErrInternalServerError  = errors.New("internal server error")
	ErrAlreadyExist         = errors.New("already exist")
	ErrNameAlreadyExist     = errors.New("name already exist")
	ErrUsernameAlreadyExist = errors.New("username already exist")
	ErrEmailAlreadyExist    = errors.New("email already exist")
	ErrPhoneAlreadyExist    = errors.New("this phone number already exist")
	ErrInvalidCursor        = errors.New("cursor is invalid")
	ErrOTPNumberNotEqual    = errors.New("otp number is not equal")
	ErrStatusNotAllow       = errors.New("status not allow")
	ErrStatusNotFound       = errors.New("not found")
	ErrBadRequest           = errors.New("bad request")
	ErrDuplicateKey         = errors.New("duplicate key")
)

var StatusInvalidENUM = func() *status.Status {
	s, _ := status.New(codes.InvalidArgument, "invalid_enum").
		WithDetails(&edpb.ErrorInfo{
			Reason: "INVALID_ENUM",
			Domain: "e-doc",
		})
	return s
}()

var StatusInvalidUUID = func() *status.Status {
	s, _ := status.New(codes.InvalidArgument, "invalid_uuid").
		WithDetails(&edpb.ErrorInfo{
			Reason: "INVALID_UUID",
			Domain: "e-doc",
		})
	return s
}()

var StatusBindingFailure = func() *status.Status {
	s, _ := status.New(codes.InvalidArgument, "binding_json_body_failure_please_pass_a_valid_json_body").
		WithDetails(&edpb.ErrorInfo{
			Reason: "BINDING_FAILURE",
			Domain: "e-doc",
		})
	return s
}()

var StatusInvalidPassword = func() *status.Status {
	s, _ := status.New(codes.InvalidArgument, "password_length_should_greater_than_6").
		WithDetails(&edpb.ErrorInfo{
			Reason: "INVALID_PASSWORD",
			Domain: "e-doc",
		})
	return s
}()

var StatusUnauthenticated = func() *status.Status {
	s, _ := status.New(codes.Unauthenticated, "id_token_not_valid_please_pass_a_valid_id_token").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "TOKEN_INVALID",
				Domain: "e-doc.com.la",
				Metadata: map[string]string{
					"service": "e-doc",
				},
			})
	return s
}()

var StatusSessionExpired = func() *status.Status {
	s, _ := status.New(codes.Unauthenticated, "session_has_been_expired_please_make_a_new session_and_try_again.").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "SESSION_EXPIRED",
				Domain: "e-doc",
			})
	return s
}()

var StatusPermissionDenied = func() *status.Status {
	s, _ := status.New(codes.PermissionDenied, "you_does't_have_sufficient_permission_to_perform_action").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INSUFFICIENT_PERMISSION",
				Domain: "e-doc",
			})
	return s
}()

var StatusNoInfo = func() *status.Status {
	s, _ := status.New(codes.NotFound, "info_not_found").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "NOT_FOUND",
				Domain: "e-doc",
			})
	return s
}()

var StatusDuplicateKey = func() *status.Status {
	s, _ := status.New(codes.InvalidArgument, "duplicate_key").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "DUPLICATE_KEY",
				Domain: "e-doc",
			})
	return s
}()

var StatusOutOfRange = func() *status.Status {
	s, _ := status.New(codes.OutOfRange, "unprocessable_entity").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "UNPROCESSABLE_ENTITY",
				Domain: "e-doc",
			})
	return s
}()

var StatusInternalServerError = func() *status.Status {
	s, _ := status.New(codes.Internal, "internal_server_error").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INTERNAL_SERVER_ERROR",
				Domain: "e-doc",
			})
	return s
}()

var StatusAlreadyExist = func() *status.Status {
	s, _ := status.New(codes.AlreadyExists, "already_exists").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "ALREADY_EXISTS",
				Domain: "e-doc",
			})
	return s
}()

var StatusNameAlreadyExist = func() *status.Status {
	s, _ := status.New(codes.AlreadyExists, "name_already_exists").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "name",
				Domain: "e-doc",
			})
	return s
}()

var StatusUsernameAlreadyExist = func() *status.Status {
	s, _ := status.New(codes.AlreadyExists, "username_already_exists").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "username",
				Domain: "e-doc",
			})
	return s
}()

var StatusPhoneAlreadyExist = func() *status.Status {
	s, _ := status.New(codes.AlreadyExists, "phone_number_already_exists").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "phone",
				Domain: "e-doc",
			})
	return s
}()

var StatusEmailAlreadyExist = func() *status.Status {
	s, _ := status.New(codes.AlreadyExists, "email_already_exists").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "email",
				Domain: "e-doc",
			})
	return s
}()

var StatusInvalidCursor = func() *status.Status {
	s, _ := status.New(codes.OutOfRange, "cursor_is_invalid").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INVALID_CURSOR",
				Domain: "e-doc",
			})
	return s
}()

var StatusOTPNumberNotEqual = func() *status.Status {
	s, _ := status.New(codes.OutOfRange, "otp_number_is_not_equal").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INVALID_OTP_NUMBER",
				Domain: "e-doc",
			})
	return s
}()

var StatusNotAllow = func() *status.Status {
	s, _ := status.New(codes.OutOfRange, "status_not_allow").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INVALID_STATUS",
				Domain: "e-doc",
			})
	return s
}()

var StatusBadRequest = func() *status.Status {
	s, _ := status.New(codes.InvalidArgument, "Invalid input. Please pass a valid values.").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INVALID_INPUT",
				Domain: "e-doc",
			})
	return s
}()

func GRPCStatusFromErr(err error) *status.Status {
	switch {
	case err == nil:
		return status.New(codes.OK, "OK")
	case errors.Is(err, ErrUnauthorized):
		return StatusUnauthenticated
	case errors.Is(err, ErrPermissionDenied):
		return StatusPermissionDenied
	case errors.Is(err, ErrBadRequest):
		return StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		return StatusUnauthenticated
	case errors.Is(err, ErrDuplicateKey):
		return StatusDuplicateKey
	case errors.Is(err, ErrUnProcessAbleEntity):
		return StatusOutOfRange
	case errors.Is(err, ErrInternalServerError):
		return StatusInternalServerError
	case errors.Is(err, ErrAlreadyExist):
		return StatusAlreadyExist
	case errors.Is(err, ErrNameAlreadyExist):
		return StatusNameAlreadyExist
	case errors.Is(err, ErrUsernameAlreadyExist):
		return StatusUsernameAlreadyExist
	case errors.Is(err, ErrEmailAlreadyExist):
		return StatusEmailAlreadyExist
	case errors.Is(err, ErrPhoneAlreadyExist):
		return StatusPhoneAlreadyExist
	case errors.Is(err, ErrInvalidCursor):
		return StatusInvalidCursor
	case errors.Is(err, ErrOTPNumberNotEqual):
		return StatusOTPNumberNotEqual
	case errors.Is(err, ErrStatusNotAllow):
		return StatusNotAllow
	}

	return StatusInternalServerError
}

func HttpStatusPbFromRPC(s *status.Status) *hspb.Error {
	return &hspb.Error{
		Error: &hspb.Error_Status{
			Code:    int32(httpStatusFromCode(s.Code())),
			Status:  code.Code(s.Code()),
			Message: s.Message(),
			Details: s.Proto().Details,
		},
	}
}

// httpStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func httpStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}
