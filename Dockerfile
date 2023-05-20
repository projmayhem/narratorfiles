FROM golang:1.20 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /narratorfiles ./cmd/narratorfiles/


FROM alpine:latest

RUN apk --no-cache add ca-certificates
COPY --from=build /narratorfiles /narratorfiles
CMD ["/narratorfiles"]
