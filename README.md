# 実験Aベンチマークサーバ
[![deploy](https://github.com/ohkilab/expA-benchmarkserver/actions/workflows/main.yml/badge.svg)](https://github.com/ohkilab/expA-benchmarkserver/actions/workflows/main.yml)

https://ohkilab.github.io/SU-CSexpA-benchmark/index.html
## 導入から起動
### Dockerを使用する場合（編集中）
1. Dockerをインストール
2. リポジトリをクローン<br>
   `$git clone git@github.com:ohkilab/expA-benchmarkserver.git`
3. クローンしたリポジトリに移動<br>
   `$cd SU-CSexpA`
4. githubに登録済みの.sshをコピー<br>
   例：`$cp -r ~/.ssh .`
   > **注意** <br>
   公開鍵と秘密鍵とknown_hostsが必要です
5. GitのEmailとユーザ名を設定する<br>
   Dockerfile内の以下を修正<br>
   `RUN git config --global user.email "hoge@hoge.co.jp"`<br>
   `RUN git config --global user.name "hoge"`<br>
6. Dockerを起動する<br>
   `$docker-compose up`
7. `http://localhost:3000`または`http://<ipアドレス>:3000`でアクセス
   
### Dockerを使用しない場合
1. [公式サイト](https://go.dev/dl/)からgoをダウンロード
   > `$go version`でgoの存在確認
3. ターミナルを再起動
4. リポジトリをクローン<br>
   `$git clone git@github.com:ohkilab/expA-benchmarkserver.git`
5. `$go run main.go`でベンチマークサーバを起動
   > **注意** <br>
   main.goプログラムが存在するディレクトリ(benchmarkserver)をカレントディレクトリにする必要があります<br>
   `$cd benchmarkserver`でmaing.goをカレントディレクトリにする
6. `http://localhost:3000`または`http://<ipアドレス>:3000`でアクセス
