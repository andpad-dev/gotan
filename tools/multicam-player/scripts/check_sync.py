"""変換後のファイル同士のずれを、指定した区間で測る. 0 に近ければ同期できている.

usage: check_sync.py <開始秒> <終了秒> <基準 wav> <wav>...
"""

import sys

import numpy as np
from scipy.io import wavfile
from scipy.signal import fftconvolve

SR = 16000
MAX_LAG = 0.5
a0, a1 = float(sys.argv[1]), float(sys.argv[2])
ref_path, *paths = sys.argv[3:]
_, ref = wavfile.read(ref_path)
ref = ref.astype(np.float32)
for p in paths:
    _, x = wavfile.read(p)
    x = x.astype(np.float32)
    a = ref[int(a0 * SR) : int(a1 * SR)]
    b = x[int((a0 - MAX_LAG) * SR) : int((a1 + MAX_LAG) * SR)]
    c = fftconvolve(b, a[::-1], mode="valid")
    print(
        f"{a0:.0f}-{a1:.0f}s\t{p}\t{(np.argmax(np.abs(c)) / SR - MAX_LAG) * 1000:+.1f} ms"
    )
