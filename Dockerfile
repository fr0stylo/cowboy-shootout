FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY . .

RUN apk add make && \
  make build

FROM alpine

WORKDIR /app

COPY --from=builder /app/bin/starter .
COPY --from=builder /app/bin/shooter .
COPY --from=builder /app/shooters.json .

ENV EXECUTABLE_PATH=./shooter

RUN ls -la .

ENTRYPOINT [ "/app/starter" ]
CMD [ "-shooters", "./shooters.json" ]
