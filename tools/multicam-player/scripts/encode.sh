#!/usr/bin/env bash
# 3 本を同期・カットして再生用ファイルを作る.
# 残す区間を wide の時刻で並べて渡すと、3 本とも同じ区間を切り出してつなぐ.
# usage: encode.sh <出力ディレクトリ> <開始秒> <終了秒> [<開始秒> <終了秒> ...]
set -euo pipefail

readonly OUT="${1:?出力ディレクトリ}"
shift
readonly SEGS=("$@")
(( ${#SEGS[@]} >= 2 && ${#SEGS[@]} % 2 == 0 )) || { echo "区間は開始と終了の組で渡す" >&2; exit 1; }

# wide の時刻 T に対応する player のコンテナ上の時刻は T + オフセット.
# 手拍子の立ち上がりから算出し、player の音声ストリームの開始遅れ (start_time) を足してある.
readonly OFF_P1="-8.9198"
readonly OFF_P2="-9.0646"
# wide は青緑に寄っているので、grayworld の結果から求めた固定の倍率で白を合わせる.
readonly WIDE_COLOR="colorchannelmixer=rr=1.738:gg=0.953:bb=0.813,"

# bc は 0.5 を .5 と出力し ffmpeg が受け付けないので、printf で整形する.
calc() { printf "%.4f" "$(echo "$1" | bc -l)"; }

mkdir -p "${OUT}"

# encode <入力> <オフセット> <高さ> <追加の映像フィルタ> <出力>
# 入力側の -ss で最初の区間の手前まで素早く飛び、-copyts で元の時刻を保ったまま trim で正確に切る.
# 画面キャプチャは可変フレームレートで切り出し位置にフレームがないことがあるので、
# STARTPTS ではなく切り出し位置を引き、fps で隙間を埋める.
encode() {
    local in="$1" off="$2" height="$3" extra="$4" out="$5"
    local first pre graph="" labels="" n=0 i s e
    first="$(calc "${SEGS[0]} + ${off}")"
    pre="$(calc "if (${first} > 10) ${first} - 10 else 0")"
    for (( i = 0; i < ${#SEGS[@]}; i += 2 )); do
        s="$(calc "${SEGS[i]} + ${off}")"
        e="$(calc "${SEGS[i + 1]} + ${off}")"
        # fps で埋めた映像が区間より長くなると、concat が音声に無音を足してずれるので、区間の長さで切りそろえる.
        graph+="[0:v:0]trim=start=${s}:end=${e},setpts=PTS-${s}/TB,fps=30:start_time=0,trim=end=$(calc "${e} - ${s}")[v${n}];"
        graph+="[0:a:0]atrim=start=${s}:end=${e},asetpts=PTS-${s}/TB[a${n}];"
        labels+="[v${n}][a${n}]"
        n=$(( n + 1 ))
    done
    graph+="${labels}concat=n=${n}:v=1:a=1[cv][a];[cv]${extra}scale=-2:${height}[v]"
    ffmpeg -hide_banner -loglevel error -y -ss "${pre}" -copyts -i "${in}" \
        -filter_complex "${graph}" \
        -map "[v]" -map "[a]" \
        -c:v libx264 -preset veryfast -crf 23 -pix_fmt yuv420p -g 60 \
        -c:a aac -b:a 128k -ac 2 \
        -movflags +faststart "${out}"
}

encode src/wide.mp4 0 1080 "${WIDE_COLOR}" "${OUT}/wide.mp4" &
encode src/player1.mov "${OFF_P1}" 1440 "" "${OUT}/player1.mp4" &
encode src/player2.mov "${OFF_P2}" 1440 "" "${OUT}/player2.mp4" &
wait
ls -l "${OUT}"
