FROM golang:alpine AS build
WORKDIR /app
COPY . .
RUN apk update && apk add build-base
RUN export CGO_ENABLED=1 && go build -o frate-template . 

FROM alpine AS final
WORKDIR /app
RUN apk update && apk add sqlite-dev  
COPY --from=build /app/frate-template /app
EXPOSE 8080
CMD ["./frate-template"]

