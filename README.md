# CLI-tool for ConoHa VPS.

[ConoHa VPS](https://www.conoha.jp/)を操作するためのCLIツールです。

ConoHa VPSのコントロールパネルは非常に使いやすく、VPSの操作に関して困ることはありません。
ただAPIなどが用意されていないため、プログラムなどからVPSを操作することは難しいです。

このツールを使うと、コマンドラインからConoHa VPSを操作することができます。
シェルスクリプトやスクリプト言語から利用することで、VPSに対してある程度の自動化を行うことが可能です。

## Features

現在対応しているのは以下の機能です。

* ログイン/ログアウト
* VPSの一覧取得
* VPSの詳細を取得(名前、プラン、IPアドレス、収容先など)
* VPSの作成
* VPSの削除
* SSH秘密鍵のダウンロード
* VPSへのSSH接続

## ビルド方法

CentOS系の場合、以下のようにします。DebianやUbuntuはyumをaptにすれば大丈夫だと思います。


```
# Goの環境をインストールします。
# また、依存ライブラリをgo getするのにMercurial(hg)が必要です。
yum install go hg

# GOPATHを設定します
export GOPATH=$HOME/go

# ソースコードを取得します
go get github.com/hironobu-s/conoha-vps

# ビルドします
cd $GOPATH/src/github.com/hironobu-s/conoha-vps
go build -o conoha

# 実行します
./conoha version
```

## インストール
(これから書く)

## クイックスタート

最初にログインします
```
$ conoha login
```

VPS一覧を取得します
```
$ conoha list
```


(つづく)

## 使い方

コマンドに -h を付けるとヘルプが表示されます。
```
$ conoha -h
```

```
$ conoha stat -h
```

(これから書く)

## TODO

* たくさん

## License

MIT License
