package handler

import (
	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/shared/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validationErrorResponse(reason, domain string, metadata map[string]string) error {
	st := status.New(codes.InvalidArgument, "invalid parameters")
	errInfo := &errdetails.ErrorInfo{
		Reason:   reason,
		Domain:   domain,
		Metadata: metadata,
	}
	st, err := st.WithDetails(errInfo)
	if err != nil {
		return status.Errorf(codes.Internal, "could not add details to the error: %v", err)
	}
	return st.Err()
}

func logRegValidationErrorResponse(v *validator.Validator) (*user.LogRegResponse, error) {
	err := validationErrorResponse("Validation", "user", v.Errors)
	return nil, err
}
