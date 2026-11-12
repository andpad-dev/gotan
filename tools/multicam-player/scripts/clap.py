"""手拍子の立ち上がりをサンプル単位で求める.

usage: clap.py <wav> <おおよその秒> [<wav> <秒> ...]
"""

import sys

import numpy as np
from scipy.io import wavfile

args = sys.argv[1:]
for path, approx in zip(args[::2], args[1::2]):
    sr, x = wavfile.read(path)
    x = np.abs(x.astype(np.float32) / 32768.0)
    s = int((float(approx) - 0.3) * sr)
    seg = x[s : s + int(0.6 * sr)]
    # 区間最大値の 30% を初めて超えたサンプルを立ち上がりとする.
    onset = int(np.argmax(seg > seg.max() * 0.3))
    print(f"{path}\t{(s + onset) / sr:.4f}")
