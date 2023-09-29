# syntax=docker/dockerfile:1
FROM golang:1.21 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /bin/snapshotter ./cmd/snapshotter
RUN chmod +x /bin/snapshotter

FROM busybox:1.36
COPY --from=build /bin/snapshotter /snapshotter
ENTRYPOINT [ "/snapshotter" ]
