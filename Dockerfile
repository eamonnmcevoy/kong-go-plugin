FROM golang:1.17.9-alpine3.15 AS builder

WORKDIR /tmp/go-plugins
COPY ./plugins /tmp/go-plugins

RUN cd /tmp/go-plugins/hello_world && go build -o hello_world .

FROM kong:2.8.1-alpine

USER root
RUN apk update && apk add gettext
WORKDIR /kong
RUN mkdir go-plugins/

COPY --from=builder /tmp/go-plugins/hello_world/hello_world /kong/go-plugins/

COPY kong.conf kong.conf ./

RUN touch kong.yaml
RUN chown -R kong:kong kong.yaml

COPY kong-tmpl.yaml .
CMD envsubst < kong-tmpl.yaml > kong.yaml && /docker-entrypoint.sh kong docker-start -c kong.conf
