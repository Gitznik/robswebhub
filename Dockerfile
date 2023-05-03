FROM rust as builder
WORKDIR /usr/src/app

COPY . .
RUN cargo build --release --bin robswebhub


FROM debian:bullseye-slim AS runtime
WORKDIR /usr/src/app
RUN apt-get update -y \
  && apt-get install -y --no-install-recommends openssl ca-certificates \
  && apt-get autoremove -y \
  && apt-get clean -y \
  && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/src/app/target/release/robswebhub robswebhub
COPY configuration configuration
ENTRYPOINT [ "./robswebhub" ]
