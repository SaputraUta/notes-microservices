# --- build stage ---
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
# go mod tidy fetches deps + writes go.sum; CGO off = static binary (pgx is pure Go)
RUN go mod tidy && CGO_ENABLED=0 go build -o /notes-backend .

# --- run stage ---
# distroless = no shell, no package manager; tiny + smaller attack surface
FROM gcr.io/distroless/static-debian12
COPY --from=build /notes-backend /notes-backend
EXPOSE 8080
ENTRYPOINT ["/notes-backend"]
