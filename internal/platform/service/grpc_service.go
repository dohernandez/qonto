package service

import (
	"context"
	"errors"

	"github.com/dohernandez/qonto/internal/domain/usecase"
	"github.com/dohernandez/qonto/internal/platform/storage"
	api "github.com/dohernandez/qonto/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// QontoService is the server that manages transfers.
type QontoService struct {
	transactionBulk usecase.TransactionBulk

	api.UnimplementedQontoServiceServer
}

// NewQontoService creates an instance of QontoService.
func NewQontoService(transactionBulk usecase.TransactionBulk) *QontoService {
	return &QontoService{
		transactionBulk: transactionBulk,
	}
}

// TransferBulk performs given transfers.
//
// Receives a request with bulk of transfer to perform. Responses whether the transfer were done successfully or not, due to:
// - account not found
// - not enough funds in the account
// - internal server.
func (s *QontoService) TransferBulk(ctx context.Context, req *api.TransferBulkRequest) (*emptypb.Empty, error) {
	input := usecase.TransactionBulkInput{
		OrganizationName: req.OrganizationName,
		OrganizationIban: req.OrganizationIban,
		OrganizationBic:  req.OrganizationBic,
	}

	input.CreditTransfers = make([]usecase.TransactionBulkTransferInput, len(req.CreditTransfers))

	for i, transfer := range req.CreditTransfers {
		input.CreditTransfers[i] = usecase.TransactionBulkTransferInput{
			Amount:           transfer.Amount,
			Currency:         transfer.Currency,
			CounterpartyName: transfer.CounterpartyName,
			CounterpartyBic:  transfer.CounterpartyBic,
			CounterpartyIban: transfer.CounterpartyIban,
			Description:      transfer.Description,
		}
	}

	err := s.transactionBulk.TransactionBulk(ctx, input)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "bank account not found")
		}

		if errors.Is(err, storage.ErrNotEnoughBalance) {
			return nil, status.Errorf(codes.FailedPrecondition, "bank account not enough balance")
		}

		return nil, status.Errorf(codes.Internal, "cannot process the transaction: %v", err)
	}

	_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "201")) // nolint: errcheck

	return &emptypb.Empty{}, nil
}
