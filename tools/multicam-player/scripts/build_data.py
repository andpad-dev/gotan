"""チャプター案 (wide の時刻) を再生用ファイルの時刻に直し、data.js を書き出す.

文字起こしの区切りは話し始めの位置なので、見る人が文脈をつかめるように
チャプターを LEAD 秒だけ前倒しする. 一覧の表示、シークバーの印、ジャンプ先は同じ時刻になる.

usage: build_data.py <chapters.json> <out data.js> <開始秒> <終了秒> [<開始秒> <終了秒> ...]
example: python3 scripts/build_data.py work/chapters.draft.json data.js 11.5 284.0 297.0 3235.5
"""

import json
import sys
from pathlib import Path

# チャプターを前倒しする秒数.
LEAD = 3.0

src, out, *segs = sys.argv[1:]
segs = [(float(a), float(b)) for a, b in zip(segs[::2], segs[1::2])]


def to_output(t):
    """wide の時刻を、残した区間をつないだ時刻に直す. 切った区間内の時刻は次の区間の頭に寄せる."""
    acc = 0.0
    for a, b in segs:
        if t < b:
            return acc + max(0.0, t - a)
        acc += b - a
    return acc


chapters = [
    {"t": round(max(0.0, to_output(c["t"]) - LEAD), 1), "title": c["title"]}
    for c in json.loads(Path(src).read_text())
]
data = {
    "title": "Go探無比 プレイ映像: slog.Value はなぜ == で比較できないのか",
    # 動画を URL で置くときはここに書く (例: "https://example.com/gotan/"). 空なら media/ から読む.
    "mediaBase": "",
    "videos": [
        {"id": "player1", "label": "プレイヤー 1 の画面", "file": "player1.mp4"},
        {"id": "player2", "label": "プレイヤー 2 の画面", "file": "player2.mp4"},
        {"id": "wide", "label": "2 人の様子", "file": "wide.mp4"},
    ],
    # ワイプに映す映像.
    "wipe": "wide",
    "defaultAudio": "wide",
    "offsets": {"player1": 0, "player2": 0, "wide": 0},
    "chapters": chapters,
}
with open(out, "w") as f:
    f.write(
        "// プレイヤーに渡すデータ. file:// で開いても読めるように、JSON ではなく JS として置く.\n"
    )
    f.write(
        "// scripts/build_data.py で生成する. 再生用ファイルの段階で同期済みなので offsets は 0.\n"
    )
    f.write("// ?dev=1 で微調整した値は offsets に書き戻す.\n")
    f.write("window.GOTAN = " + json.dumps(data, ensure_ascii=False, indent=2) + ";\n")
for c in chapters:
    print(f"{int(c['t'] // 60):2d}:{c['t'] % 60:04.1f}  {c['title']}")
