# syntax=docker/dockerfile:1

FROM node:24-bookworm-slim AS web-builder
WORKDIR /build/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM rust:1.97-bookworm AS rust-builder
WORKDIR /build
COPY Cargo.toml Cargo.lock build.rs rust-toolchain.toml ./
COPY proto/ proto/
COPY migrations/ migrations/
COPY src/ src/
RUN --mount=type=cache,id=cyberedge-cargo-registry,target=/usr/local/cargo/registry \
    --mount=type=cache,id=cyberedge-cargo-git,target=/usr/local/cargo/git \
    --mount=type=cache,id=cyberedge-cargo-target,target=/build/target \
    cargo build --release --bins \
    && mkdir -p /out \
    && cp target/release/cyberedge target/release/cyberedge-agent /out/

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /opt/cyberedge
COPY --from=rust-builder /out/cyberedge /usr/local/bin/cyberedge
COPY --from=rust-builder /out/cyberedge-agent /usr/local/bin/cyberedge-agent
COPY --from=web-builder /build/web/dist ./web
USER 65532:65532
EXPOSE 50051 8080
ENTRYPOINT ["cyberedge"]
