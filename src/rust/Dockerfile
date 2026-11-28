# Byggkontexten är src/ - samma som Go-tjänstens Dockerfile.
#
#   docker build -f rust/Dockerfile -t plan-api-rust src/
#
# Stage 1 bygger binären, stage 2 är distroless med bara binären.
# OpenAPI-specen genereras ur koden (utoipa), så ingen fil behöver
# kopieras in - till skillnad från Go-tjänsten som laddar swagger.json
# från disk.
#
# Alla workspace-medlemmars Cargo.toml kopieras in före koden, så att
# beroendena cachas separat. Cargo kräver att varje medlem i workspace-
# manifestet finns, även under cachebygget.

FROM rust:1.99-alpine AS build
WORKDIR /build

COPY rust/Cargo.toml rust/Cargo.lock ./
COPY rust/crates/plan-api/Cargo.toml crates/plan-api/Cargo.toml
COPY rust/crates/plan-agent/Cargo.toml crates/plan-agent/Cargo.toml
COPY rust/crates/plan-parity/Cargo.toml crates/plan-parity/Cargo.toml
COPY rust/crates/plan-recurrence/Cargo.toml crates/plan-recurrence/Cargo.toml
COPY rust/crates/plan-wiki/Cargo.toml crates/plan-wiki/Cargo.toml

# Cargo kräver att varje deklarerat mål finns, även under cachebygget.
# Placeholders skapas utifrån manifestens path-fält, så att en ny bin
# inte behöver handpåläggas här.
COPY rust/docker/placeholders.sh ./placeholders.sh
RUN sh ./placeholders.sh && cargo fetch --locked

COPY rust/crates ./crates
RUN cargo build --release --locked --bin plan-api

FROM gcr.io/distroless/static-debian12:nonroot AS production
WORKDIR /app

COPY --from=build /build/target/release/plan-api ./plan-api

ENV PORT=8080
EXPOSE 8080

# Vilken build som körs. Sätts av byggworkflowen (--build-arg) så att
# en smoke-test efter rollout kan se att den nya versionen svarar.
ARG BUILD_SHA=unknown
ENV PLAN_BUILD_SHA=${BUILD_SHA}

# Distroless saknar curl/wget - samma binär gör TCP-proben.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/plan-api", "healthcheck"]

USER nonroot:nonroot

ENTRYPOINT ["/app/plan-api"]
