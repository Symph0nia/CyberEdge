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
    && cp target/release/cyberedge target/release/cyberedge-agent target/release/cyberedge-renderer target/release/cyberedge-nuclei-adapter /out/

FROM debian:bookworm-slim AS renderer
RUN apt-get -o Acquire::Retries=5 update \
    && renderer_attempt=0 \
    && until apt-get -o Acquire::Retries=5 install -y --no-install-recommends ca-certificates chromium; do \
         renderer_attempt=$((renderer_attempt + 1)); \
         [ "$renderer_attempt" -lt 5 ] || exit 1; \
         sleep 2; \
       done \
    && rm -rf /var/lib/apt/lists/*
RUN install -d -o 65532 -g 65532 /run/cyberedge-renderer
COPY --from=rust-builder /out/cyberedge-renderer /usr/local/bin/cyberedge-renderer
USER 65532:65532
ENTRYPOINT ["cyberedge-renderer"]

FROM projectdiscovery/nuclei:v3.11.0@sha256:e677842fb1f50f29747565ba274a1d35dcf8c684132a42b0cb406e71fccae9fc AS nuclei-upstream

FROM debian:bookworm-slim AS nuclei-adapter
RUN apt-get -o Acquire::Retries=5 update \
    && apt-get -o Acquire::Retries=5 install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && install -d -o 65532 -g 65532 /run/cyberedge-nuclei
COPY --from=nuclei-upstream /usr/local/bin/nuclei /usr/local/bin/nuclei
COPY --from=rust-builder /out/cyberedge-nuclei-adapter /usr/local/bin/cyberedge-nuclei-adapter
USER 65532:65532
ENTRYPOINT ["cyberedge-nuclei-adapter"]

FROM debian:bookworm-slim AS core
RUN apt-get -o Acquire::Retries=5 update \
    && apt-get -o Acquire::Retries=5 install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN install -d -o 65532 -g 65532 /run/cyberedge-renderer
RUN install -d -o 65532 -g 65532 /run/cyberedge-nuclei
WORKDIR /opt/cyberedge
COPY --from=rust-builder /out/cyberedge /usr/local/bin/cyberedge
COPY --from=rust-builder /out/cyberedge-agent /usr/local/bin/cyberedge-agent
COPY --from=web-builder /build/web/dist ./web
USER 65532:65532
EXPOSE 50051 8080
ENTRYPOINT ["cyberedge"]
