# Stage 1: Build
FROM docker.io/golang:1.23.3 AS builder

ENV GOARCH=amd64

WORKDIR /root

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o scully .

# Stage 2: Final minimal image
FROM docker.io/ubuntu:22.04

ARG USERNAME=scully
ARG USERID=1231

WORKDIR /app

COPY --from=builder /root/scully ./

RUN apt-get update && apt-get install -y libc6 && rm -rf /var/lib/apt/lists/* && \
    groupadd --gid $USERID $USERNAME && \
    useradd --uid $USERID --gid $USERID --no-create-home --shell /sbin/nologin $USERNAME && \
    chown -R $USERNAME:$USERNAME /app

USER $USERNAME

CMD ["./scully"]
