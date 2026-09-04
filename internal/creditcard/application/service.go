package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	carddomain "github.com/lucas/financial-api/internal/creditcard/domain"
)

var (
	ErrNotFound             = errors.New("credit card not found")
	ErrAlreadyExists        = errors.New("credit card already exists")
	ErrArchived             = errors.New("credit card is archived")
	ErrInstitutionNotFound  = errors.New("financial institution not found")
	ErrInstitutionArchived  = errors.New("financial institution is archived")
	ErrPeriodOverlap        = errors.New("credit card period overlaps another period")
	ErrConfigurationMissing = errors.New("credit card configuration is missing")
)

type Repository interface {
	Create(context.Context, string, string, string, carddomain.ConfigurationInput) (carddomain.Card, error)
	List(context.Context, string, *carddomain.Status) ([]carddomain.Card, error)
	Find(context.Context, string, string) (carddomain.Card, error)
	UpdateName(context.Context, string, string, string) (carddomain.Card, error)
	ChangeConfiguration(context.Context, string, string, carddomain.ConfigurationInput) (carddomain.Card, error)
	Archive(context.Context, string, string) (carddomain.Card, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

type CreateInput struct {
	InstitutionID string
	Name          string
	Configuration carddomain.ConfigurationInput
}

type ListFilters struct{ Status *carddomain.Status }
type UpdateInput struct{ Name string }
type ChangeInput struct{ Configuration carddomain.ConfigurationInput }
type ArchiveInput struct{ Reason *string }

func (service *Service) Create(ctx context.Context, ownerID string, input CreateInput) (carddomain.Card, error) {
	if err := validateIdentity(ownerID, input.InstitutionID, "institution"); err != nil {
		return carddomain.Card{}, err
	}
	name, err := carddomain.NormalizeName(input.Name)
	if err != nil {
		return carddomain.Card{}, err
	}
	if err := carddomain.ValidateConfiguration(input.Configuration); err != nil {
		return carddomain.Card{}, err
	}
	return service.repository.Create(ctx, ownerID, input.InstitutionID, name, input.Configuration)
}

func (service *Service) List(ctx context.Context, ownerID string, filters ListFilters) ([]carddomain.Card, error) {
	if err := validateOwner(ownerID); err != nil {
		return nil, err
	}
	if filters.Status != nil && !filters.Status.Valid() {
		return nil, carddomain.ErrValidation
	}
	return service.repository.List(ctx, ownerID, filters.Status)
}

func (service *Service) Find(ctx context.Context, ownerID, cardID string) (carddomain.Card, error) {
	if err := validateIdentity(ownerID, cardID, "card"); err != nil {
		return carddomain.Card{}, err
	}
	return service.repository.Find(ctx, ownerID, cardID)
}

func (service *Service) Update(ctx context.Context, ownerID, cardID string, input UpdateInput) (carddomain.Card, error) {
	if err := validateIdentity(ownerID, cardID, "card"); err != nil {
		return carddomain.Card{}, err
	}
	name, err := carddomain.NormalizeName(input.Name)
	if err != nil {
		return carddomain.Card{}, err
	}
	return service.repository.UpdateName(ctx, ownerID, cardID, name)
}

func (service *Service) Change(ctx context.Context, ownerID, cardID string, input ChangeInput) (carddomain.Card, error) {
	if err := validateIdentity(ownerID, cardID, "card"); err != nil {
		return carddomain.Card{}, err
	}
	if err := carddomain.ValidateConfiguration(input.Configuration); err != nil {
		return carddomain.Card{}, err
	}
	return service.repository.ChangeConfiguration(ctx, ownerID, cardID, input.Configuration)
}

func (service *Service) Archive(ctx context.Context, ownerID, cardID string, input ArchiveInput) (carddomain.Card, error) {
	if err := validateIdentity(ownerID, cardID, "card"); err != nil {
		return carddomain.Card{}, err
	}
	if err := carddomain.ValidateReason(input.Reason); err != nil {
		return carddomain.Card{}, err
	}
	return service.repository.Archive(ctx, ownerID, cardID)
}

func validateIdentity(ownerID, id, resource string) error {
	if err := validateOwner(ownerID); err != nil {
		return err
	}
	if !carddomain.ValidID(id) {
		return fmt.Errorf("%w: %s id must be a UUID", carddomain.ErrValidation, resource)
	}
	return nil
}

func validateOwner(ownerID string) error {
	if strings.TrimSpace(ownerID) == "" {
		return fmt.Errorf("%w: authenticated user is required", carddomain.ErrValidation)
	}
	return nil
}
