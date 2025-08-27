package email

import (
	"context"
	"mizu/internal/email/params"
)

type EmailSender interface {
	Init()
	SendVerifyRequest(ctx context.Context, arg *params.VerifyRequest) error
	SendContractSigned(ctx context.Context, arg *params.ContractSignedRequest) error
	SendContractRejected(ctx context.Context, arg *params.ContractRejectedRequest) error
	Close()
}
