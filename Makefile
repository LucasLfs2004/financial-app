.PHONY: run test test-race format supabase-start supabase-stop db-reset db-reset-local db-push db-push-cloud

run:
	go run ./cmd/api

test:
	go test ./...

test-race:
	go test -race ./...

format:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

supabase-start:
	supabase start

supabase-stop:
	supabase stop

db-reset:
	$(MAKE) db-reset-local

db-reset-local:
	supabase db reset --local --yes

db-push:
	@echo "Use 'make db-push-cloud' para publicar migrations no projeto vinculado."
	@exit 1

db-push-cloud:
	supabase db push --linked --dry-run
	supabase db push --linked
	supabase migration list --linked
