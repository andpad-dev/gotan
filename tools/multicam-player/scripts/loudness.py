"""30 秒ごとの音量 (dBFS) を並べ、収録の頭と尻で人が話しているかを見る.

usage: loudness.py <wav>...
"""

import sys

import numpy as np
from scipy.io import wavfile

win = 30
cols = []
for p in sys.argv[1:]:
    sr, x = wavfile.read(p)
    x = x.astype(np.float32) / 32768.0
    n = len(x) // (sr * win)
    rms = np.sqrt((x[: n * sr * win].reshape(n, -1) ** 2).mean(axis=1))
    cols.append(20 * np.log10(rms + 1e-9))
m = min(len(c) for c in cols)
print(
    "time   "
    + "  ".join(p.split("/")[-1].split(".")[0][:7].rjust(7) for p in sys.argv[1:])
)
for i in range(m):
    print(
        f"{i * win // 60:3d}:{i * win % 60:02d} "
        + "  ".join(f"{c[i]:7.1f}" for c in cols)
    )
