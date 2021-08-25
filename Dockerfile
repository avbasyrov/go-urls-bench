# Сборка приложения
FROM golang:latest as builder
RUN mkdir /build
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o app ./main.go


# Для рантайма используем только исполняемый файл
FROM alpine:latest
RUN mkdir /runtime
WORKDIR /runtime
COPY --from=builder /build/app /runtime/app
COPY --from=builder /build/config.yml /runtime/config.yml

ENTRYPOINT [ "/runtime/app" ]
# CMD [ "1", "2", "3" ]
