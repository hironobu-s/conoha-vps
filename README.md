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

GOPATH配下にgit cloneして

```
go build
```

## インストール
## クイックスタート
## 使い方

(これから書く)

## TODO

* たくさん

## License

MIT License
