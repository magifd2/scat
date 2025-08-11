# scat: 汎用コマンドラインコンテンツ投稿ツール

`scat` は、ファイルや標準入力から受け取ったコンテンツを、Slackなどの設定済み宛先に送信するための、多機能なコマンドラインインターフェースです。`slackcat` にインスパイアされていますが、より汎用的で拡張しやすいように設計されています。

[![Build Status](https://github.com/magifd2/scat/actions/workflows/build.yml/badge.svg)](https://github.com/magifd2/scat/actions/workflows/build.yml)

---

## 主な機能

- **テキストメッセージの投稿**: 引数、ファイル、標準入力からコンテンツを送信します。
- **ファイルのアップロード**: 指定したパスや標準入力からファイルをアップロードします。
- **コンテンツのストリーミング**: 標準入力を継続的に監視し、定期的にメッセージを投稿します。
- **プロファイル管理**: 複数の宛先を設定し、簡単に切り替えることができます。
- **拡張可能なプロバイダ**: 現在、Slackとテスト用のモックプロバイダをサポートしています。

## インストール

[リリースページ](https://github.com/magifd2/scat/releases)から、お使いのシステム用の最新のバイナリをダウンロードしてください。

または、ソースからビルドすることも可能です:

```bash
make build
```

## 初期セットアップ

投稿を開始する前に、設定ファイルを作成する必要があります。

1.  **設定ファイルの初期化**:

    以下のコマンドを実行して、デフォルトの場所に設定ファイル (`~/.config/scat/config.json`) を作成します:

    ```bash
    scat config init
    ```

2.  **プロファイルの設定**:

    デフォルトのプロファイルは、テストに便利なモックプロバイダを使用します。Slackのような実際のサービスに投稿するには、新しいプロファイルを追加する必要があります。

    Slackプロファイル設定の詳細な手順については、**[Slackセットアップガイド](./SLACK_SETUP.md)** を参照してください。

    以下に、新しいSlackプロファイルを簡単に追加する例を示します:

    ```bash
    # このコマンドを実行すると、Slack Botトークンを安全に入力するよう求められます。
    scat profile add my-slack-workspace --provider slack --channel "#general"
    ```

3.  **アクティブプロファイルの設定**:

    `scat` がデフォルトで使用するプロファイルを指定します:

    ```bash
    scat profile use my-slack-workspace
    ```

## 使用例

`scat` の一般的な使い方をいくつか紹介します。

### テキストメッセージの投稿 (`post`)

-   **引数から投稿**:

    ```bash
    scat post "コマンドラインからこんにちは！"
    ```

-   **標準入力から (パイプ)**:

    ```bash
    echo "このメッセージはパイプされました。" | scat post
    ```

-   **ファイルから投稿**:

    ```bash
    scat post --from-file ./message.txt
    ```

-   **標準入力からストリーミング**:

    ログの監視などに便利です。`scat` は入力をバッファリングし、数秒ごとにまとめて投稿します。

    ```bash
    tail -f /var/log/system.log | scat post --stream
    ```

### ファイルのアップロード (`upload`)

-   **パスを指定してファイルをアップロード**:

    ```bash
    scat upload --file ./report.pdf
    ```

-   **コメント付きでアップロード**:

    ```bash
    scat upload --file ./screenshot.png -m "こちらがご依頼のスクリーンショットです。"
    ```

-   **標準入力からアップロード**:

    標準入力からストリーミングする場合、アップロード用のファイル名を指定する必要があります。

    ```bash
    cat data.csv | scat upload --file - --filename data.csv
    ```

### プロファイル管理 (`profile`)

-   **すべてのプロファイルを表示**:

    アクティブなプロファイルにはアスタリスク (`*`) が付きます。

    ```bash
    scat profile list
    ```

-   **アクティブなプロファイルを切り替え**:

    ```bash
    scat profile use another-profile
    ```

-   **特定のプロファイルでコマンドを実行** (アクティブプロファイルは変更しない):

    ```bash
    scat --profile personal-slack post "これは個人用ワークスペースへのメッセージです。"
    ```

-   **プロファイル設定の追加・変更**:

    ```bash
    scat profile set channel "#random"
    ```

### グローバルフラグ

これらのフラグは、どのコマンドでも使用できます。

-   `--debug`: 詳細なデバッグログを有効にします。
-   `--silent`: 成功メッセージを抑制します。
-   `--noop`: 実際にコンテンツを投稿・アップロードしないドライランを実行します。
-   `--profile <name>`: コマンドの実行に特定のプロファイルを使用します。