## ビルド方法

以下のコマンドを実行

```sh
GOOS=linux go build -o bin/main
```

## デプロイ方法

```
sls deploy
```

## 試しに実行する場合

```bash
sls invoke -f billing-notification
```