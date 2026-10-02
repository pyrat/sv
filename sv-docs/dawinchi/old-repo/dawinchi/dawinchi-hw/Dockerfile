FROM golang:1.21.1 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s" -v -o /winchi-hw cmd/winch/main.go

# Now copy it into our base image.
FROM gcr.io/distroless/base
COPY --from=build /winchi-hw /winchi-hw

EXPOSE 9001

CMD ["/winchi-hw"]
