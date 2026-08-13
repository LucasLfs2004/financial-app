.PHONY: run test test-race format supabase-start supabase-stop db-reset db-push

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
	supabase db reset

db-push:
	supabase db push
