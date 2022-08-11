FROM node:16.13.2-alpine as builder-node
LABEL stage=builder
WORKDIR /usr/src/app

COPY package.json /usr/src/app/package.json
COPY package-lock.json /usr/src/app/package-lock.json
RUN npm install
COPY . /usr/src/app
RUN npm run build

FROM golang:1.18-alpine as builder-go
LABEL stage=builder
WORKDIR /usr/src/app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
 apk add --no-cache upx ca-certificates tzdata
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
COPY --from=builder-node /usr/src/app/dist /usr/src/app/frontimport/static
RUN CGO_ENABLED=0 go build -tags prod -ldflags "-s -w" -o registry ./cmd/app &&\
 upx --best registry -o _upx_registry && \
 mv -f _upx_registry registry

FROM scratch as runner
COPY --from=builder-go /usr/share/zoneinfo/Asia/Yekaterinburg /etc/localtime
COPY --from=builder-go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder-go /usr/src/app/registry /opt/app/
CMD ["/opt/app/registry"]