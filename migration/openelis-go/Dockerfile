# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /src
# No external deps yet (go.mod only, no go.sum). Copy source and build a static
# binary so the runtime image can stay minimal.
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/openelis ./cmd/openelis

# ---- runtime stage ----
FROM alpine:3.20
# tzdata so Go can resolve IANA zones (parity: ZoneId.systemDefault().getId()).
RUN apk add --no-cache tzdata ca-certificates
COPY --from=build /out/openelis /usr/local/bin/openelis
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/openelis"]
