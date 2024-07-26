package auth

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
	ErrUnauthorized        = errors.New("unauthorized")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrNoInfo              = errors.New("no info")
	ErrUnProcessAbleEntity = errors.New("unprocessable entity")
	ErrInternalServerError = errors.New("internal server error")
)

var StatusBindingFailure = func() *status.Status {
	s, _ := status.New(codes.InvalidArgument, "binding_json_body_failure_please_pass_a_valid_json_body").
		WithDetails(&edpb.ErrorInfo{
			Reason: "BINDING_FAILURE",
			Domain: "htmx",
		})
	return s
}()

var StatusUnauthenticated = func() *status.Status {
	s, _ := status.New(codes.Unauthenticated, "id_token_not_valid_please_pass_a_valid_id_token").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "TOKEN_INVALID",
				Domain: "htmx.com.la",
				Metadata: map[string]string{
					"service": "htmx",
				},
			})
	return s
}()

var StatusSessionExpired = func() *status.Status {
	s, _ := status.New(codes.Unauthenticated, "session_has_been_expired_please_make_a_new session_and_try_again.").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "SESSION_EXPIRED",
				Domain: "htmx",
			})
	return s
}()

var StatusPermissionDenied = func() *status.Status {
	s, _ := status.New(codes.PermissionDenied, "you_does't_have_sufficient_permission_to_perform_action").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INSUFFICIENT_PERMISSION",
				Domain: "htmx",
			})
	return s
}()

var StatusNoInfo = func() *status.Status {
	s, _ := status.New(codes.NotFound, "info_not_found").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "NOT_FOUND",
				Domain: "htmx",
			})
	return s
}()

var StatusOutOfRange = func() *status.Status {
	s, _ := status.New(codes.OutOfRange, "unprocessable_entity").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "UNPROCESSABLE_ENTITY",
				Domain: "htmx",
			})
	return s
}()

var StatusInternalServerError = func() *status.Status {
	s, _ := status.New(codes.Internal, "internal_server_error").
		WithDetails(
			&edpb.ErrorInfo{
				Reason: "INTERNAL_SERVER_ERROR",
				Domain: "htmx",
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
	case errors.Is(err, ErrNoInfo):
		return StatusNoInfo
	case errors.Is(err, ErrUnProcessAbleEntity):
		return StatusOutOfRange
	case errors.Is(err, ErrInternalServerError):
		return StatusInternalServerError
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
