package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
)

var (
	ErrNotFound       = errors.New("financial institution not found")
	ErrAlreadyExists  = errors.New("financial institution already exists")
	ErrArchived       = errors.New("financial institution is archived")
	ErrHasActiveCards = errors.New("financial institution has active cards")
)

type Repository interface {
	Create(context.Context, string, string) (institutiondomain.Institution, error)
	List(context.Context, string, *institutiondomain.Status) ([]institutiondomain.Institution, error)
	UpdateName(context.Context, string, string, string) (institutiondomain.Institution, error)
	Archive(context.Context, string, string) (institutiondomain.Institution, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

type CreateInput struct {
	Name string
}

type ListFilters struct {
	Status *institutiondomain.Status
}

type UpdateInput struct {
	Name string
}

type ArchiveInput struct {
	Reason *string
}

func (service *Service) Create(ctx context.Context, ownerID string, input CreateInput) (institutiondomain.Institution, error) {
	if err := validateOwner(ownerID); err != nil {
		return institutiondomain.Institution{}, err
	}
	name, err := institutiondomain.NormalizeName(input.Name)
	if err != nil {
		return institutiondomain.Institution{}, err
	}
	return service.repository.Create(ctx, ownerID, name)
}

func (service *Service) List(ctx context.Context, ownerID string, filters ListFilters) ([]institutiondomain.Institution, error) {
	if err := validateOwner(ownerID); err != nil {
		return nil, err
	}
	if filters.Status != nil && !filters.Status.Valid() {
		return nil, institutiondomain.ErrValidation
	}
	return service.repository.List(ctx, ownerID, filters.Status)
}

func (service *Service) Update(ctx context.Context, ownerID, institutionID string, input UpdateInput) (institutiondomain.Institution, error) {
	if err := validateIdentity(ownerID, institutionID); err != nil {
		return institutiondomain.Institution{}, err
	}
	name, err := institutiondomain.NormalizeName(input.Name)
	if err != nil {
		return institutiondomain.Institution{}, err
	}
	return service.repository.UpdateName(ctx, ownerID, institutionID, name)
}

func (service *Service) Archive(ctx context.Context, ownerID, institutionID string, input ArchiveInput) (institutiondomain.Institution, error) {
	if err := validateIdentity(ownerID, institutionID); err != nil {
		return institutiondomain.Institution{}, err
	}
	if err := institutiondomain.ValidateReason(input.Reason); err != nil {
		return institutiondomain.Institution{}, err
	}
	return service.repository.Archive(ctx, ownerID, institutionID)
}

func validateIdentity(ownerID, institutionID string) error {
	if err := validateOwner(ownerID); err != nil {
		return err
	}
	if !institutiondomain.ValidID(institutionID) {
		return fmt.Errorf("%w: institution id must be a UUID", institutiondomain.ErrValidation)
	}
	return nil
}

func validateOwner(ownerID string) error {
	if strings.TrimSpace(ownerID) == "" {
		return fmt.Errorf("%w: authenticated user is required", institutiondomain.ErrValidation)
	}
	return nil
}
