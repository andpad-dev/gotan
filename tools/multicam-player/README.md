# multicam-player

同じ場面を撮った複数の動画（プレイヤー 2 人の画面キャプチャと、2 人を写した引きの映像）を、同期して再生する 1 枚の HTML です。
ワークショップのプレイ映像を振り返るために作りました。

動画はサイズが大きいので、リポジトリには入れていません。

## 見る

1. 再生用の動画 `player1.mp4`、`player2.mp4`、`wide.mp4` を用意する。
2. このディレクトリに `media/` を作り、3 本をそこに置く。
3. `index.html` をブラウザで開く。

再生ページは、次の順で動画を探します。

1. `data.js` の `mediaBase` に書いた URL（今は空）
2. `index.html` と同じ階層の `media/`
3. どちらにもなければ、フォルダ選択ボタンを出す。3 本が入ったフォルダを選ぶと再生できる

動画を URL で置く場合は、`cmd/build-data/main.go` の `MediaBase` に URL を書いて `data.js` を作り直します。

### 操作

| キー | 操作 |
| --- | --- |
| Space | 再生 / 停止 |
| ← → | 10 秒戻る / 進む（Shift を押しながらだと 60 秒） |
| 1 2 | 1 画面 / 2 画面 |
| W | ワイプ（2 人の様子の小窓）を出す / 隠す |
| M | ミュート / 解除 |
| C | チャプター一覧を開く / 閉じる |
| F | 全画面 / 解除 |
| [ ] | 速度を下げる / 上げる（0.5〜2 倍） |

ワイプはドラッグすると、四隅と中央下のうち近い位置に移ります。
ワイプにマウスを重ねると、大きさ（小・中・大）を選べます。
チャプター一覧の開閉と、ワイプの位置と大きさは、ブラウザに保存されます。

`?dev=1` を付けて開くと、1 フレーム単位で同期を微調整する欄が出ます。
調整した値は `data.js` の `offsets` に書き戻します。

## 作り直す

元の動画にはタイムコードがないので、冒頭の手拍子の音で 3 本の時刻を合わせています。
コマンドはすべてこのディレクトリで実行します。
Go 1.26 以上と ffmpeg を使います。
文字起こしだけは、mlx-whisper を使うために uv と Apple Silicon の Mac が要ります。

| ディレクトリ | 中身 | コミット |
| --- | --- | --- |
| `src/` | 元の動画（`wide.mp4`、`player1.mov`、`player2.mov`）へのリンク | しない |
| `work/` | 16kHz の音声、文字起こし、チャプター案などの中間ファイル | しない |
| `media/` | 再生用の動画 | しない |
| `cmd/`、`internal/` | 以下の手順の Go のコマンド | する |
| `scripts/` | ffmpeg を呼ぶシェルスクリプトと、文字起こしの Python スクリプト | する |

1. 音声を取り出し、ずれ量を測る。

   ```sh
   for f in wide player1 player2; do
     ffmpeg -i src/$f.* -vn -ac 1 -ar 16000 work/$f.16k.wav
   done
   go run ./cmd/sync work/wide.16k.wav work/player1.16k.wav work/player2.16k.wav
   # sync が出した手拍子の候補 (おおよその秒) を渡し、立ち上がりをサンプル単位で求める
   go run ./cmd/clap work/wide.16k.wav 10.3 work/player1.16k.wav 1.3 work/player2.16k.wav 1.2
   ```

   「player の手拍子の時刻 − wide の手拍子の時刻」に、player の音声の開始遅れ（`ffprobe` の `start_time`）を足した値を、`scripts/encode.sh` の `OFF_P1` と `OFF_P2` に書く。

2. 残す区間を wide の時刻で指定し、3 本を同期・カットして変換する。

   ```sh
   scripts/encode.sh media 11.5 284.0 297.0 3235.5
   ```

3. 変換後のずれを確かめる（序盤・中盤・終盤で ±45 ms 以内だった）。

   ```sh
   mkdir -p work/final
   for f in wide player1 player2; do
     ffmpeg -i media/$f.mp4 -vn -ac 1 -ar 16000 work/final/$f.wav
   done
   go run ./cmd/check-sync 60 120 work/final/wide.wav work/final/player1.wav work/final/player2.wav
   ```

4. 文字起こしをして、チャプター案を作る。

   ```sh
   # 各マイクは主にその席の人の声を拾うので、player 2 本を文字起こしして合わせる
   # (mlx-whisper のコマンド版は温度のフォールバックを指定できず、同じ文の繰り返しが出るので、Python から呼ぶ)
   for f in player1 player2; do
     uv run --with mlx-whisper scripts/transcribe.py work/$f.16k.wav work/$f.transcript.json
   done
   go run ./cmd/merge-transcript work/transcript.tsv
   ```

   `work/transcript.tsv` を読んで、`work/chapters.draft.json` に wide の時刻でチャプターを書く。

5. チャプターを再生用の時刻に直して `data.js` を書き出す。
   チャプターは、話し始めの 3 秒前に置く（`cmd/build-data` の `lead`）。

   ```sh
   go run ./cmd/build-data work/chapters.draft.json data.js 11.5 284.0 297.0 3235.5
   ```

6. オフラインで配る zip を作る。

   ```sh
   scripts/package.sh
   ```

手順の途中で使う補助のコマンドもあります。

| コマンド | 用途 |
| --- | --- |
| `go run ./cmd/loudness <wav>...` | 30 秒ごとの音量を並べ、収録の頭と尻で人が話しているかを見る |
| `go run ./cmd/quiet <wav> <開始秒> <終了秒>` | 0.5 秒ごとの音量を出し、カットに向いた静かな切れ目を探す |
| `go run ./cmd/wb-gains src/wide.mp4 <秒>...` | wide の色補正の倍率を求める（`scripts/encode.sh` の `WIDE_COLOR`） |

動作確認でシークを試すときは、`file://` でも開けます。
HTTP で開く場合は、Range リクエストに対応した `cmd/serve` を使います（`python3 -m http.server` ではシークが先頭に戻る）。

```sh
go run ./cmd/serve . 8765
```
