FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -o bin/main cmd/main.go

FROM debian:12
RUN apt-get update && apt-get install -y ca-certificates \
    && apt-get clean
WORKDIR /root/
COPY --from=builder /app/bin/main /root/
COPY --from=builder /app/assets /root/assets
RUN chmod +x /root/main
# # Install Doppler CLI
# RUN wget -q -t3 'https://packages.doppler.com/public/cli/rsa.8004D9FF50437357.key' -O /etc/apk/keys/cli@doppler-8004D9FF50437357.rsa.pub && \
#     echo 'https://packages.doppler.com/public/cli/alpine/any-version/main' | tee -a /etc/apk/repositories && \
#     apk add doppler

# CMD ["doppler", "run", "--","./main"]
CMD ["./main"]
