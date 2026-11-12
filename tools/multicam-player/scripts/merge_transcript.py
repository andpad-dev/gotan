"""player1 と player2 の文字起こしを wide の時刻軸に揃えて 1 本にする.

player の音声の t 秒は wide の t + (手拍子の時刻差) 秒にあたる.
usage: merge_transcript.py <out.tsv>
"""

import json
import sys
from pathlib import Path

SHIFT = {"P1": 10.3101 - 1.2606, "P2": 10.3101 - 1.1711}
rows = []
for tag, path in [
    ("P1", "work/player1.transcript.json"),
    ("P2", "work/player2.transcript.json"),
]:
    for s in json.loads(Path(path).read_text()):
        rows.append((s["start"] + SHIFT[tag], tag, s["text"].strip()))
rows.sort()
with open(sys.argv[1], "w") as f:
    for t, tag, text in rows:
        f.write(f"{int(t // 60):02d}:{t % 60:05.2f}\t{tag}\t{text}\n")
print(len(rows))
