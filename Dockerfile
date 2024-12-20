FROM golang:alpine AS build
WORKDIR /app
RUN apk update && apk add build-base
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" export CGO_ENABLED=1 && go build -o frate-template

FROM alpine AS final
WORKDIR /app
RUN apk update && apk add sqlite-dev  
COPY --from=build /app/frate-template /app
EXPOSE 8080
COPY templates.db /app
CMD ["./frate-template"]

