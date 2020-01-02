## ビルド方法

以下のコマンドを実行

```sh
GOARCH=amd64 GOOS=linux go build
```

## デプロイ方法

以下のコマンドで実行ファイルをzip化
※ zip化する前にbuildすること

```sh
zip billing-notification.zip ./billing-notification
```

AWS ConsoleからLambdaを選択して、任意の関数パッケージをアップロードする
