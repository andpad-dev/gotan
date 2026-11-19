"""mlx-whisper で文字起こしし、セグメントを JSON に書き出す.

usage: transcribe.py <wav> <out.json>
"""

import json
import sys

import mlx_whisper

wav, out = sys.argv[1:]
r = mlx_whisper.transcribe(
    wav,
    path_or_hf_repo="mlx-community/whisper-large-v3-turbo",
    language="ja",
    condition_on_previous_text=False,
)
segs = [
    {"start": round(s["start"], 2), "end": round(s["end"], 2), "text": s["text"]}
    for s in r["segments"]
]
with open(out, "w") as f:
    json.dump(segs, f, ensure_ascii=False, indent=1)
