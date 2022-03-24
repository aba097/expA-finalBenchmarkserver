#goのベース環境をもってくる
FROM golang:1.17-alpine as builder

#アップデートとgitのインストール
RUN apk update && apk add git alpine-sdk

#abコマンドのインストール
RUN apk --no-cache add apache2-utils

#sshコマンドのインストール
RUN apk add openssh

#タイムゾーンの設定
RUN apk --update add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    apk del tzdata && \
    rm -rf /var/cache/apk/*

#sshキーの追加
ADD .ssh /root/.ssh
RUN chmod 600 /root/.ssh*

#githubからクローン
WORKDIR /go/src
#githubからclone
RUN git clone git@github.com:ohkilab/SU-CSexpA-benchmark.git
RUN git config --global user.email "hoge@hoge.co.jp"
RUN git config --global user.name "hoge"

#ベンチマークサーバを起動
WORKDIR /go/src/SU-CSexpA-benchmark/benchmarkserver
CMD ["go","run","main.go"]
