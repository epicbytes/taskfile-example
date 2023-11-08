FROM node:18-alpine as builder-frontend
WORKDIR /build
COPY frontend .
RUN npm install
RUN npm run build

FROM golang:1.20-alpine as builder-backend
WORKDIR /build
COPY backend .
RUN go mod download
COPY --from=builder-frontend build/dist dist
RUN CGO_ENABLED=0 go build -o /main main.go

FROM scratch
ENV ENVIRONMENT=production
COPY --from=builder-backend main /bin/main
EXPOSE 8099
ENTRYPOINT ["/bin/main"]
