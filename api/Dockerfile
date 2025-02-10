ARG GO_VERSION=1.23.6
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo,osusergo -ldflags="-w -s -extldflags '-static'" -v -o web-api .


FROM gcr.io/distroless/static-debian12

EXPOSE 3001
ENV PORT 3001

COPY --from=builder /go/src/app/web-api /
COPY --from=builder /go/src/app/internal/email /email
CMD ["/web-api"]
