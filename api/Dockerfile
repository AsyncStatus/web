ARG GO_VERSION=1.23.6
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo,osusergo -ldflags="-w -s -extldflags '-static'" -v -o api .
RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo,osusergo -ldflags="-w -s -extldflags '-static'" -v -o migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12

EXPOSE 3001
ENV PORT=3001

COPY --from=builder /go/src/app/api /api
COPY --from=builder /go/src/app/migrate /migrate
COPY --from=builder /go/src/app/migrations /migrations
COPY --from=builder /go/src/app/internal/email /internal/email
CMD ["/api"]
