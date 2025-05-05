FROM golang:1.24 AS build

# Support to pull dependencies from private repositories
#RUN git config --global url.ssh://git@github.com/.insteadOf https://github.com/
#RUN mkdir -p /root/.ssh && chmod 0700 /root/.ssh && ssh-keyscan github.com > /root/.ssh/known_hosts
#COPY id_rsa /root/.ssh/id_rsa
#RUN chmod 600 /root/.ssh/id_rsa && ssh-keyscan github.com >> /root/.ssh/known_hosts

WORKDIR /go/src/app

COPY . .

ARG APP_TAG_VERSION=unknown-dev

RUN ls -alF
RUN go mod tidy
RUN CGO_ENABLED=0 go build \
    -ldflags "-X 'github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/apiserver.AppTagVersion=${APP_TAG_VERSION}'" \
    -o /go/bin/app ./main.go

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/src/app/configs /configs
COPY --from=build /go/src/app/dbmigrate/ /dbmigrate/
COPY --from=build /go/bin/app /app

EXPOSE 8080

CMD ["/app", "run", "--deployment=dev"]
