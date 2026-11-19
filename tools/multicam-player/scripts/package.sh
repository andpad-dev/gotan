#!/usr/bin/env bash
# 再生ページと動画をまとめた、オフラインで見られる配布用 zip を作る.
# 動画は圧縮済みなので、zip では圧縮せず格納だけにする (-0).
# usage: scripts/package.sh [<出力 zip>]   (tools/multicam-player で実行する)
set -euo pipefail

readonly OUT="${1:-gotan-play.zip}"
readonly NAME="gotan-play"
readonly FILES=(player1.mp4 player2.mp4 wide.mp4)

for f in "${FILES[@]}"; do
    [[ -f "media/${f}" ]] || { echo "media/${f} がない. 先に scripts/encode.sh で作る" >&2; exit 1; }
done

STAGE="$(mktemp -d)"
trap 'rm -rf "${STAGE}"' EXIT
mkdir -p "${STAGE}/${NAME}/media"
cp index.html data.js "${STAGE}/${NAME}/"
# 動画はコピーせずリンクで置く. zip はリンク先の中身を格納する.
for f in "${FILES[@]}"; do ln -s "${PWD}/media/${f}" "${STAGE}/${NAME}/media/${f}"; done
cat > "${STAGE}/${NAME}/README.txt" <<'EOF'
Go探無比 プレイ映像

index.html をダブルクリックし、ブラウザ（Chrome / Edge / Safari / Firefox）で開いてください。
media フォルダは index.html と同じ場所に置いたままにしてください。
うまく再生できないときは Chrome で開いてください。

3 本の映像（2 人の画面と、2 人の様子）は同期して再生されます。

操作
  Space            再生 / 停止
  ← →              10 秒戻る / 進む
  Shift + ← →      60 秒戻る / 進む
  1 2              1 画面 / 2 画面
  W                ワイプ（2 人の様子の小窓）を出す / 隠す
  M                ミュート / 解除
  C                チャプター一覧を開く / 閉じる
  F                全画面 / 解除
  [ ]              速度を下げる / 上げる (0.5〜2 倍)
  ワイプはドラッグで四隅か中央下へ動かせます。マウスを重ねると大きさを選べます。
  右側のチャプターをクリックすると、その場面へ移動します。
EOF

rm -f "${OUT}"
(cd "${STAGE}" && zip -0 -r -q "${OLDPWD}/${OUT}" "${NAME}")
ls -lh "${OUT}"
