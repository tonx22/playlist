FROM golang:1.19 as builder
WORKDIR /
COPY . ./playlist
WORKDIR /playlist
RUN CGO_ENABLED=0 GOOS=linux go build -a -o cloud-app
#CMD [ "./cloud-app" ]

FROM alpine:3.16
WORKDIR /playlist
COPY --from=builder /playlist/cloud-app .
COPY --from=builder /playlist/migrations ./migrations
CMD [ "./cloud-app" ]