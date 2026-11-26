# Byggkontexten är src/ - samma som Go-tjänstens Dockerfile.
#
#   docker build -f rust/Dockerfile -t plan-api-rust src/
#
# Stage 1 bygger binären, stage 2 är distroless med bara binären.
# OpenAPI-specen genereras ur koden (utoipa), så ingen fil behöver
# kopieras in - till skillnad från Go-tjänsten som laddar swagger.json
# från disk.

FROM rust:1.99-alpine AS build
WORKDIR /build

# Beroendena cachas separat från koden, så att kodändringar inte laddar
# ner crates igen. placeholder-main.rs behövs eftersom Cargo kräver ett
# mål i varje gång.
COPY rust/Cargo.toml rust/Cargo.lock ./
COPY rust/crates/plan-api/Cargo.toml crates/plan-api/Cargo.toml
RUN mkdir -p crates/plan-api/src \
    && echo 'fn main() {}' > crates/plan-api/src/main.rs \
    && cargo build --release --locked \
    && rm -rf crates/plan-api/src

COPY rust/crates ./crates
RUN touch crates/plan-api/src/main.rs \
    && cargo build --release --locked

FROM gcr.io/distroless/static-debian12:nonroot AS production
WORKDIR /app

COPY --from=build /build/target/release/plan-api ./plan-api

ENV PORT=8080
EXPOSE 8080

# Distroless saknar curl/wget - samma binär gör TCP-proben.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/plan-api", "healthcheck"]

USER nonroot:nonroot

ENTRYPOINT ["/app/plan-api"]
