FROM alpine:3.16

RUN apk --no-cache upgrade && apk --no-cache add ca-certificates

COPY ./twiyou /app

WORKDIR /app

CMD ["bin/sh", "-c", "echo Running against DB ${DB_NAME} && /app/twiyou"]
