package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	institutionapplication "github.com/lucas/financial-api/internal/financialinstitution/application"
	institutiondomain "github.com/lucas/financial-api/internal/financialinstitution/domain"
	"github.com/lucas/financial-api/internal/platform/auth"
	"github.com/lucas/financial-api/internal/platform/httpapi"
)

const (
	testOwnerID       = "a1000000-0000-0000-0000-000000000001"
	testInstitutionID = "a1100000-0000-0000-0000-000000000001"
)

type fakeService struct {
	createdInput institutionapplication.CreateInput
	listFilters  institutionapplication.ListFilters
	updatedID    string
	archivedID   string
	err          error
}

func (service *fakeService) Create(_ context.Context, _ string, input institutionapplication.CreateInput) (institutiondomain.Institution, error) {
	service.createdInput = input
	return sampleInstitution(), service.err
}

func (service *fakeService) List(_ context.Context, _ string, filters institutionapplication.ListFilters) ([]institutiondomain.Institution, error) {
	service.listFilters = filters
	return []institutiondomain.Institution{sampleInstitution()}, service.err
}

func (service *fakeService) Update(_ context.Context, _, institutionID string, _ institutionapplication.UpdateInput) (institutiondomain.Institution, error) {
	service.updatedID = institutionID
	return sampleInstitution(), service.err
}

func (service *fakeService) Archive(_ context.Context, _, institutionID string, _ institutionapplication.ArchiveInput) (institutiondomain.Institution, error) {
	service.archivedID = institutionID
	institution := sampleInstitution()
	institution.Status = institutiondomain.StatusArchived
	return institution, service.err
}

func TestRoutesCreateAndListInstitutions(t *testing.T) {
	service := &fakeService{}
	mux := http.NewServeMux()
	RegisterRoutes(mux, authenticated, service)

	created := performRequest(t, mux, http.MethodPost, "/v1/financial-institutions", `{"name":"Banco Alfa"}`)
	if created.Code != http.StatusCreated || service.createdInput.Name != "Banco Alfa" {
		t.Fatalf("status=%d input=%+v body=%s", created.Code, service.createdInput, created.Body.String())
	}
	assertInstitutionEnvelope(t, created)

	listed := performRequest(t, mux, http.MethodGet, "/v1/financial-institutions?status=active", "")
	if listed.Code != http.StatusOK || service.listFilters.Status == nil || *service.listFilters.Status != institutiondomain.StatusActive {
		t.Fatalf("status=%d filters=%+v body=%s", listed.Code, service.listFilters, listed.Body.String())
	}
}

func TestRoutesUpdateAndArchiveInstitution(t *testing.T) {
	service := &fakeService{}
	mux := http.NewServeMux()
	RegisterRoutes(mux, authenticated, service)

	updated := performRequest(t, mux, http.MethodPatch, "/v1/financial-institutions/"+testInstitutionID, `{"name":"Banco Beta"}`)
	if updated.Code != http.StatusOK || service.updatedID != testInstitutionID {
		t.Fatalf("status=%d id=%q body=%s", updated.Code, service.updatedID, updated.Body.String())
	}

	archived := performRequest(t, mux, http.MethodPost, "/v1/financial-institutions/"+testInstitutionID+"/archive", `{"reason":"antiga"}`)
	if archived.Code != http.StatusOK || service.archivedID != testInstitutionID {
		t.Fatalf("status=%d id=%q body=%s", archived.Code, service.archivedID, archived.Body.String())
	}
}

func TestRoutesMapValidationAndConflictErrors(t *testing.T) {
	service := &fakeService{}
	mux := http.NewServeMux()
	RegisterRoutes(mux, authenticated, service)

	invalid := performRequest(t, mux, http.MethodPost, "/v1/financial-institutions", `{"unknown":true}`)
	assertErrorCode(t, invalid, http.StatusBadRequest, "validation_error")

	service.err = institutionapplication.ErrHasActiveCards
	conflict := performRequest(t, mux, http.MethodPost, "/v1/financial-institutions/"+testInstitutionID+"/archive", "")
	assertErrorCode(t, conflict, http.StatusConflict, "institution_has_active_cards")
}

func authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		ctx := httpapi.WithPrincipal(request.Context(), auth.Principal{UserID: testOwnerID})
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

func performRequest(t *testing.T, handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func sampleInstitution() institutiondomain.Institution {
	now := time.Date(2026, time.September, 3, 12, 0, 0, 0, time.UTC)
	return institutiondomain.Institution{ID: testInstitutionID, UserID: testOwnerID, Name: "Banco Alfa", Status: institutiondomain.StatusActive, CreatedAt: now, UpdatedAt: now}
}

func assertInstitutionEnvelope(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data["id"] != testInstitutionID || envelope.Data["name"] != "Banco Alfa" {
		t.Fatalf("data=%+v", envelope.Data)
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if response.Code != status || envelope.Error.Code != code {
		t.Fatalf("status=%d code=%q body=%s", response.Code, envelope.Error.Code, response.Body.String())
	}
}
