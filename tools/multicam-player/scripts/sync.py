"""3 本の音声からずれ量を測る.

wide を基準に、player の音声が何秒遅れて始まっているかを求める.
冒頭の手拍子と、冒頭・末尾 10 分の相互相関の両方で測り、ドリフトの有無を確かめる.

usage: uv run --with numpy --with scipy scripts/sync.py work/wide.16k.wav work/player1.16k.wav work/player2.16k.wav
"""

import json
import sys

import numpy as np
from scipy.io import wavfile
from scipy.signal import fftconvolve

SR = 16000


def load(path):
    sr, x = wavfile.read(path)
    assert sr == SR, path
    return x.astype(np.float32) / 32768.0


def clap_candidates(x, search_sec=600, n=5):
    """冒頭 search_sec 秒から、立ち上がりの鋭い大きなピークを n 個返す (秒)."""
    seg = x[: search_sec * SR]
    hop = SR // 100  # 10ms.
    env = np.sqrt(np.convolve(seg**2, np.ones(hop) / hop, mode="same"))[::hop]
    # 直前 0.5 秒の平均との比で立ち上がりの鋭さを測る.
    bg = np.convolve(env, np.ones(50) / 50, mode="full")[: len(env)]
    score = env / (bg + 1e-4)
    out = []
    for i in np.argsort(score)[::-1]:
        if all(abs(i - j) > 100 for j, _, _ in out):
            out.append((i, float(score[i]), float(env[i])))
        if len(out) == n:
            break
    return sorted((i / 100, s, e) for i, s, e in out)


def xcorr_offset(ref, x, ref_start, x_start, length, max_lag):
    """ref[ref_start:] と x[x_start:] の区間を比べ、x が ref より何秒遅れているかを返す."""
    a = ref[int(ref_start * SR) : int((ref_start + length) * SR)]
    b = x[int((x_start - max_lag) * SR) : int((x_start + length + max_lag) * SR)]
    c = fftconvolve(b, a[::-1], mode="valid")
    k = int(np.argmax(np.abs(c)))
    peak = np.abs(c[k]) / (np.linalg.norm(a) * np.linalg.norm(b[k : k + len(a)]) + 1e-9)
    lag = (x_start - max_lag) + k / SR - ref_start
    return lag, float(peak)


def main():
    ref_path, *paths = sys.argv[1:]
    ref = load(ref_path)
    result = {"reference": ref_path, "clap": {}, "pairs": {}}
    result["clap"][ref_path] = clap_candidates(ref)
    for p in paths:
        x = load(p)
        result["clap"][p] = clap_candidates(x)
        # 粗い推定: 冒頭 10 分同士を ±60 秒の範囲で比べる.
        coarse, cpk = xcorr_offset(ref, x, 120, 120, 300, 60)
        # 末尾側: 同じずれを仮定して 10 分ほど比べ、ドリフトを測る.
        tail_ref = len(ref) / SR - 700
        tail, tpk = xcorr_offset(ref, x, tail_ref, tail_ref + coarse, 300, 5)
        result["pairs"][p] = {
            "head_offset_sec": round(coarse, 4),
            "head_peak": round(cpk, 3),
            "tail_offset_sec": round(tail, 4),
            "tail_peak": round(tpk, 3),
            "drift_sec": round(tail - coarse, 4),
            "duration_sec": len(x) / SR,
        }
    result["reference_duration_sec"] = len(ref) / SR
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
