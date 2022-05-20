#goのベース環境をもってくる
FROM golang:1.17-alpine as builder

#アップデートとgitのインストール
RUN apk update && apk add git alpine-sdk

#abコマンドのインストール
RUN apk --no-cache add apache2-utils

#bnechmarkプログラムを追加
ADD benchmarkserver /go/src/benchmarkserver

#ベンチマークサーバを起動
WORKDIR /go/src/benchmarkserver
CMD ["go","run","main.go"]
