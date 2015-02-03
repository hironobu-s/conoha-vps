
# CLI-tool for ConoHa VPS.

[![Build Status](https://travis-ci.org/hironobu-s/conoha-vps.svg?branch=master)](https://travis-ci.org/hironobu-s/conoha-vps)

[ConoHa VPS](https://www.conoha.jp/)を操作するためのCLIツールです。

ConoHa VPSのコントロールパネルは非常に使いやすく、VPSの操作に関して困ることはありません。
ただAPIなどが用意されていないため、プログラムなどからVPSを操作することは難しいです。

このツールを使うと、コマンドラインからConoHa VPSを操作することができます。
シェルスクリプトやスクリプト言語から利用することで、VPSに対してある程度の自動化を行うことが可能です。

## 特長

現在対応しているのは以下の機能です。

* ログイン/ログアウト
* VPSの一覧取得
* VPSの詳細を取得(名前、プラン、IPアドレス、収容先など)
* VPSの作成
* VPSの削除
* SSH秘密鍵のダウンロード
* VPSへのSSH接続


## インストール

### MacOSX

ターミナルなどから以下のコマンドを実行します。

```bash
L=/usr/local/bin/conoha && curl -sL https://github.com/hironobu-s/conoha-vps/releases/download/v20150204.1/conoha-osx.amd64.gz | zcat > $L && chmod +x $L
```

アンインストールする場合は/usr/local/bin/conohaを削除してください。

### Linux

ターミナルなどから以下のコマンドを実行します。/usr/local/binにインストールされるので、root権限が必要です。他のディレクトリにインストールする場合はL=/usr/local/bin/conohaの部分を適宜書き換えてください。

```bash
L=/usr/local/bin/conoha && curl -sL https://github.com/hironobu-s/conoha-vps/releases/download/v20150204.1/conoha-linux.amd64.gz | zcat > $L && chmod +x $L
```

アンインストールする場合は/usr/local/bin/conohaを削除してください。

### Windows

[ZIPファイル](https://github.com/hironobu-s/conoha-vps/releases/download/v20150204.1/conoha.amd64.zip)をダウンロードして、適当なフォルダに展開します。

実行する場合は、コマンドプロンプトから実行してください(ファイル名をダブルクリックしても何も起きません)。

アンインストールする場合はファイルをゴミ箱に入れてください。


## クイックスタート

最初にログインします
```
$ conoha login
```

すると、ConoHaアカウントの入力プロンプトが出るので入力します。アカウントが正しいと「Login Successfully」となりログイン成功です。
```
$ conoha login
Please input ConoHa accounts.
ConoHa Account: [ACCOUNT]
Password: [PASSWORD]
INFO[0004] Login Successfully.
```

ログインに成功すると、全コマンドが実行できるようになります。コマンドの一覧は-hを付けると表示されます。
```
$ conoha -h
Usage: conoha COMMAND [OPTIONS]

DESCRIPTION
    A CLI-Tool for ConoHa VPS.

COMMANDS
    login    Authenticate an account.
    list     List VPS.
    add      Add VPS.
    remove   Remove VPS.
    ssh-key  Download and store SSH Private key.
    ssh      Login to VPS via SSH.
    logout   Remove an authenticate file(~/.conoha-vps).
    version  Print version.
```

まずはlistコマンドを実行してみましょう。VPSの一覧が表示されます。
```
$conoha list -v
VPS ID                  Label                   Plan                            Server Status   Service Status          CreatedAt
f648a6646b7e7d91        CentOS7                 8GB Memory                      Running         In operation            2015/01/27 13:15 JST
a2ae45355615d641        UbuntuDesktop           4GB Memory                      Offline         In operation            2014/12/11 16:59 JST
28ff51fd97a96106        WindowsServer2012       8GB Memory  - Windows           Running         In operation            2014/11/13 10:21 JST
```

次にstatコマンドを実行してみましょう。メニューが表示され、選択したVPSの詳細情報が表示されます。(サンプルのため一部を***でマスクしています)
```
$ conoha stat
[1] CentOS7
[2] UbuntuDesktop
[3] WindowsServer2012
Please select VPS no. [1-3]: 1
VPS ID               f648a6646b7e7d91
ServerStatus         Running
Label                CentOS7
ServiceStatus        In operation
Service ID           VPS00708435
Plan                 8GB Memory
Created At           2015-01-27T13:15:00+09:00
Delete Date          0001-01-01 00:00:00 +0000 UTC
Payment Span         1month
CPU                  Virtual6Core
Memory               8192MB
Disk1                HDD 20GB
Disk2                HDD 780GB
IPv4 Address         ***.***.***.***
IPv4 Netmask         255.255.254.0
IPv4 Gateway         ***.***.***.***
IPv4 DNS1            ***.***.***.***
IPv4 DNS2            ***.***.***.***
Host Server          cnode-f0000
Common Server ID     iu3-0000000
Serial Console(SSH)  console1001.cnode.jp
ISO Upload(SFTP)     sftp1001.cnode.jp
```

このように、コマンドライン操作でVPSを操作することができます。


## コマンド一覧

conohaコマンドがサポートしている機能の一覧です。
全てのコマンドに共通で、-hオプションを付けて実行すると、使い方を表示します。

### add

新しいVPSを追加します。以下のオプションを組み合わせることで、すべてのプラン種別(標準プラン=basic、Windowsプラン=windows)、プラン(1G, 2G, 4G, 8G, 16G)、テンプレートイメージ(CentOS, Nginx+WordPressなど)に対応します。

[オプション]
* -t, --type:     VPS種別を指定します。"basic"か"windows"である必要があります。
* -p: --plan:     プランを指定します。1Gプランの場合は1，2Gプランの場合は2、と言うように指定します。
* -P: --password: rootパスワードを指定します。標準プランのみです。
* -i: --image:    テンプレートイメージを指定します。"centos" "wordpress" "windows2012" "windows2008"のどれかである必要があります。

[コマンド実行例]

標準プラン1GBのVPSを追加する場合
```
$ conoha add -t basic -p 1 -i centos -P {password}
```

標準プラン2GBのWordPressテンプレートを使ったVPSを追加する場合
```
$ conoha add -t basic -p 4 -i wordpress -P {password}
```

Windowsプランの8GBでWindows Server 2012を追加する場合
```
$ conoha add -t windows -p 8 -i windows2012
```

Windowsプランの16GBでWindows Server 2008を追加する場合
```
$ conoha add -t windows -p 16 -i windows2008
```

### list

VPSの一覧を表示します。
そのまま実行するとServer Status列が空欄になります。これはオプションで-vを付けると取得されます。

* -v, --verbose: ServerStatusを取得します。その代わり実行に少し時間がかかります。
* -i, --id-only: VPS-ID列のみを表示します。シェルスクリプトで使うときに便利です。

```
$ conoha list -v
VPS ID                  Label                   Plan                            Server Status   Service Status          CreatedAt
f648a6646b7e7d91        CentOS7                 8GB Memory                      Running         In operation            2015/01/27 13:15 JST
a2ae45355615d641        UbuntuDesktop           4GB Memory                      Offline         In operation            2014/12/11 16:59 JST
28ff51fd97a96106        WindowsServer2012       8GB Memory  - Windows           Running         In operation            2014/11/13 10:21 JST
```

### login

ConoHaアカウントでログインします。versionなど一部のコマンドを除き、コマンドの実行にはログインが必須です。
実行するとアカウントとパスワードの入力プロンプトが表示されるので入力してください。

アカウントとパスワードはオプション(-aと-p)で渡すこともできます。

> **NOTE:** アカウントとパスワードなどをファイルに保持します。ファイルはホームディレクトリの.conoha-vpsで、パーミッションは0600です。

```
$ conoha login
Please input ConoHa accounts.
ConoHa Account: [ACCOUNT]
Password: [PASSWORD]
INFO[0004] Login Successfully.
$
```


### logout

ログアウトして認証ファイルを削除します。

```
$ conoha logout
```


### remove

VPSを削除します。実行すると、本当に削除するか確認ダイアログが表示され、Yesと回答すると削除が実行されます。
複数のVPSがある場合はVPSを選択するプロンプトが表示されますが、引数でVPS-IDを直接指定することもできます。

[オプション]
* -f, --force-remove:  確認プロンプトを表示せず直ちに削除を実行します。

```
$ conoha remove
Remove VPS[Label=VPS00712702]. Are you sure?
[y/N]: y
INFO[0009] Removing VPS is complete.
```


### ssh-key

アカウントに紐付いたSSH秘密鍵を取得し保存します。
秘密鍵はconoha-[ACCOUNT].keyと言うファイル名で保存されますが、オプションでファイル名を指定することもできます。

[オプション]
* -f, --file: 保存する秘密鍵のファイル名を指定します

```
$ conoha ssh-key
INFO[0000] Download is complete. A private key is stored in "conoha-000000.key".
```

### ssh

SSHを使用してVPSに直接ログインします。
複数のVPSがある場合はVPSを選択するプロンプトが表示されますが、引数でVPS-IDを直接指定することもできます。

このサブコマンドはsshコマンドのラッパーなので、sshがサポートしている機能は全て使えます。また渡されたオプションはサブコマンドのオプションでないものは全てsshコマンドにそのまま渡されます。これはつまり以下のような使い方ができると言うことです。

```
# -i オプションで秘密鍵ファイルを指定している。
$ conoha ssh -i ~/.ssh/private.key 
```

> **NOTE:** このサブコマンドはsshクライアントがインストールされていることが前提になります。またWindows環境では動作しません。

[オプション]
* -u: --user:     SSH接続時のユーザ名を指定します。(デフォルトではrootを使用します)

```
$ conoha ssh
Warning: Permanently added '***.***.***.***' (RSA) to the list of known hosts.
Last login: Mon Feb  2 10:53:56 2015 from myhome.com
[root@v***.***.***.*** ~]#
```

### stat

VPSの表債情報を表示します。
複数のVPSがある場合はVPSを選択するプロンプトが表示されますが、引数でVPS-IDを直接指定することもできます。

オプションなしで実行すると、IPv6情報を表示しません。


[オプション]

* -6, --include-ipv6: 出力にIPv6情報を含めます

```
$ conoha stat
(省略。出力サンプルについては「クイックスタート」をご覧ください)
```

### version

バージョンを表示します。

```
$ conoha version
v20150203.4
```

## ビルド方法

自分でビルドする場合は、以下の手順を参考にしてください。
DebianやUbuntuはyumをaptにすれば大丈夫だと思います。

```
# Goの環境をインストール
# また、依存ライブラリをgo getするのにMercurial(hg)が必要です。
yum install go hg

# GOPATHを設定
export GOPATH=$HOME/go

# ソースコードを取得
go get github.com/hironobu-s/conoha-vps

# ビルド
cd $GOPATH/src/github.com/hironobu-s/conoha-vps
go build -o conoha

# 実行
./conoha version
```

## TODO

* OS再インストールのサポート
* statが遅い
* ほかいろいろ

## License

MIT License
