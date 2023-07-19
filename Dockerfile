FROM rust as builder
WORKDIR /usr/src/app

COPY . .
RUN cargo build --release --bin robswebhub

#fontconfig libfontconfig1
FROM debian:bullseye-slim AS runtime
WORKDIR /usr/src/app
RUN apt-get update -y \
  && apt-get install -y --no-install-recommends openssl ca-certificates pkg-config fontconfig libfontconfig1 libfreetype6-dev libfontconfig1-dev \
  && apt-get autoremove -y \
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/src/app/target/release/robswebhub robswebhub
COPY configuration configuration
COPY images images
ENTRYPOINT [ "./robswebhub" ]
