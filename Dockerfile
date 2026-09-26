FROM node:22-alpine AS assets
WORKDIR /src
COPY package*.json tailwind.config.js ./
RUN npm install
COPY web ./web
RUN npm run css:build

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
COPY --from=assets /src/web/static/css/app.css ./web/static/css/app.css
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /pos-phoenix ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache tzdata ca-certificates && adduser -D -u 10001 app && mkdir /data && chown app:app /data
USER app
COPY --from=build /pos-phoenix /pos-phoenix
COPY --from=build /src/web /web
ENV ADDR=:8080 DATABASE_PATH=/data/pos.db SESSION_SECURE=true
EXPOSE 8080
CMD ["/pos-phoenix"]
