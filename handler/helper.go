package handler

import (
	"github.com/glekoz/online-shop_proto/user"
	"github.com/glekoz/online-shop_user/shared/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// func validationErrorResponsehh(reason, domain string, metadata map[string]string) error {
// 	st := status.New(codes.InvalidArgument, "invalid parameters")
// 	errInfo := &errdetails.ErrorInfo{
// 		Reason:   reason,
// 		Domain:   domain,
// 		Metadata: metadata,
// 	}
// 	st, err := st.WithDetails(errInfo)
// 	if err != nil {
// 		return status.Errorf(codes.Internal, "could not add details to the error: %v", err)
// 	}
// 	return st.Err()
// }

func badRequestResponse(reason string, metadata map[string]string) error {
	if len(metadata) < 1 {
		return status.Error(codes.InvalidArgument, "Bad Request")
	}
	st := status.New(codes.InvalidArgument, reason)
	fields := make([]*errdetails.BadRequest_FieldViolation, 0, len(metadata))
	for k, v := range metadata {
		fields = append(fields, &errdetails.BadRequest_FieldViolation{
			Field:       k,
			Description: v,
		})
	}
	details := errdetails.BadRequest{
		FieldViolations: fields,
	}
	st, err := st.WithDetails(&details)
	if err != nil {
		return status.Error(codes.Internal, "детали не добавились")
	}
	return st.Err()
}

func logRegBadRequestResponse(v *validator.Validator) (*user.LogRegResponse, error) {
	err := badRequestResponse("validation of provided credentials failed", v.Errors)
	return nil, err
}
