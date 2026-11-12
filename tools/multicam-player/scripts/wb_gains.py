"""wide の色補正の倍率を求める.

複数時刻のフレームで、元画像と grayworld 補正後の RGB 平均の比を取り、その中央値を全編共通の倍率にする.
usage: wb_gains.py <動画> <秒>...
"""

import subprocess
import sys

import numpy as np

W, H = 320, 180


def frame(path, t, vf):
    raw = subprocess.run(
        [
            "ffmpeg",
            "-loglevel",
            "error",
            "-ss",
            str(t),
            "-i",
            path,
            "-frames:v",
            "1",
            "-vf",
            f"scale={W}:{H},{vf}format=rgb24",
            "-f",
            "rawvideo",
            "-",
        ],
        capture_output=True,
        check=True,
    ).stdout
    return np.frombuffer(raw, np.uint8).reshape(H, W, 3).astype(np.float64)


path, *times = sys.argv[1:]
gains = []
for t in times:
    a = frame(path, t, "")
    b = frame(path, t, "format=gbrpf32le,grayworld,")
    gains.append(b.reshape(-1, 3).mean(0) / a.reshape(-1, 3).mean(0))
g = np.median(gains, axis=0)
print("per-frame:", np.round(gains, 3).tolist())
print(f"colorchannelmixer=rr={g[0]:.3f}:gg={g[1]:.3f}:bb={g[2]:.3f}")
