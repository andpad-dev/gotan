"""区間内で 0.5 秒ごとの音量を出し、静かな切れ目を探す. usage: quiet.py <wav> <開始秒> <終了秒>"""

import sys

import numpy as np
from scipy.io import wavfile

sr, x = wavfile.read(sys.argv[1])
x = x.astype(np.float32) / 32768
a, b = float(sys.argv[2]), float(sys.argv[3])
t = a
while t < b:
    s = x[int(t * sr) : int((t + 0.5) * sr)]
    print(f"{t:7.1f} {20 * np.log10(np.sqrt((s**2).mean()) + 1e-9):6.1f}")
    t += 0.5
