#-----------------------------------------------------------------------------
FROM golang:alpine AS go-builder

RUN apk add --no-cache upx

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /go/src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# RUN ls
RUN go build \
      -trimpath \
      -ldflags="-s -w -extldflags '-static'" \
      -o /go/bin/main \
      ./cmd/meme

RUN upx --lzma /go/bin/main


#-----------------------------------------------------------------------------
FROM node:12-alpine AS elm-builder

RUN apk add --no-cache curl

WORKDIR /elm

RUN mkdir src
COPY elm/src/* src/
COPY elm/elm.json ./
RUN curl -L -o elm.gz https://github.com/elm/compiler/releases/download/0.19.1/binary-for-linux-64-bit.gz && \
  gunzip elm.gz && \
  chmod +x elm && \
  mv elm /usr/local/bin && \
  npm install -g uglify-js && \
  elm make --optimize --output=elm.js src/Main.elm && \
  uglifyjs elm.js --compress 'pure_funcs=[F2,F3,F4,F5,F6,F7,F8,F9,A2,A3,A4,A5,A6,A7,A8,A9],pure_getters,keep_fargs=false,unsafe_comps,unsafe' | uglifyjs --mangle --output elm.min.js

#-----------------------------------------------------------------------------
FROM scratch

ENV GIN_MODE=release

COPY web web
COPY --from=go-builder /go/bin/main .
COPY --from=elm-builder /elm/elm.min.js web/elm.min.js

ENTRYPOINT ["./main"]
