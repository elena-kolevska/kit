package status

import (
	customerrors "github.com/dapr/kit/pkg/proto/custom_errors/v1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ConstructError(code codes.Code, err error, key, errDescription, resourceType, reason, owner, domain, resourceName string, metadata map[string]string) error {
	ei := ConstructErrorInfo(domain, key, reason, metadata)
	ri := ConstructResourceInfo(owner, resourceName, errDescription, resourceType)
	kei := ConstructDaprKitErrorInfo(domain, key, reason, metadata)
	ste, stErr := status.Newf(code, errDescription).WithDetails(ei, ri, kei)
	if stErr != nil {
		return err
	}

	return ste.Err()
}

func ConstructErrorInfo(domain, key, reason string, metadata map[string]string) *errdetails.ErrorInfo {
	ei := errdetails.ErrorInfo{
		Domain: domain,
		Reason: reason,
		Metadata: map[string]string{
			"key": key,
		},
	}
	for k, v := range metadata {
		ei.Metadata[k] = v
	}
	return &ei
}

func ConstructDaprKitErrorInfo(domain, key, reason string, metadata map[string]string) *customerrors.DaprKitErrorInfo {
	ei := customerrors.DaprKitErrorInfo{
		Domain: domain,
		Reason: reason,
		Metadata: map[string]string{
			"key": key,
		},
	}
	for k, v := range metadata {
		ei.Metadata[k] = v
	}
	return &ei
}

func ConstructResourceInfo(owner, resourceName, description string, resourceType string) *errdetails.ResourceInfo {
	return &errdetails.ResourceInfo{
		ResourceType: resourceType,
		ResourceName: resourceName,
		Owner:        owner,
		Description:  description,
	}
}
