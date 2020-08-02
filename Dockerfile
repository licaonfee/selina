FROM golang:alpine AS builder

WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o selina cmd/*.go

FROM alpine 

WORKDIR /app
COPY --from=builder /workspace/selina .

ENTRYPOINT ["/app/selina"]
CMD [ "-help" ]
