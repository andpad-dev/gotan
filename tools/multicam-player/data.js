// プレイヤーに渡すデータ. file:// で開いても読めるように、JSON ではなく JS として置く.
// go run ./cmd/build-data で生成する. 再生用ファイルの段階で同期済みなので offsets は 0.
// ?dev=1 で微調整した値は offsets に書き戻す.
window.GOTAN = {
  "title": "Go探無比 プレイ映像: slog.Value はなぜ == で比較できないのか",
  "mediaBase": "",
  "videos": [
    {
      "id": "player1",
      "label": "プレイヤー 1 の画面",
      "file": "player1.mp4"
    },
    {
      "id": "player2",
      "label": "プレイヤー 2 の画面",
      "file": "player2.mp4"
    },
    {
      "id": "wide",
      "label": "2 人の様子",
      "file": "wide.mp4"
    }
  ],
  "wipe": "wide",
  "defaultAudio": "wide",
  "offsets": {
    "player1": 0,
    "player2": 0,
    "wide": 0
  },
  "chapters": [
    {
      "t": 0.0,
      "title": "シナリオ選びと問題文の確認"
    },
    {
      "t": 78.5,
      "title": "調査の入り口から slog パッケージへ"
    },
    {
      "t": 199.5,
      "title": "設問 1: value.go で disallow == を発見"
    },
    {
      "t": 346.0,
      "title": "設問 1: Equal メソッドの実装を読む"
    },
    {
      "t": 471.0,
      "title": "言語仕様の比較演算子の節を読む"
    },
    {
      "t": 767.5,
      "title": "設問 1 の答え合わせ"
    },
    {
      "t": 841.0,
      "title": "設問 2: ヒントを手がかりに仕様を再読"
    },
    {
      "t": 1074.0,
      "title": "設問 2 の答え合わせ"
    },
    {
      "t": 1162.0,
      "title": "設問 3: ゼロ長配列のサイズを測る"
    },
    {
      "t": 1350.0,
      "title": "設問 3: 仕様のサイズとアラインメント保証"
    },
    {
      "t": 1559.5,
      "title": "設問 3: コンパイラで zero-sized field を検索"
    },
    {
      "t": 1732.0,
      "title": "設問 3: Issue #9401 で経緯を読む"
    },
    {
      "t": 2056.5,
      "title": "設問 3 の答え合わせ"
    },
    {
      "t": 2178.0,
      "title": "設問 4: Playground で比較を試す"
    },
    {
      "t": 2317.5,
      "title": "設問 4: 仕様の interface 比較の項を読む"
    },
    {
      "t": 2457.0,
      "title": "設問 4: 履歴から CL と Issue をたどる"
    },
    {
      "t": 2749.0,
      "title": "設問 4 の答えを確認"
    },
    {
      "t": 2803.0,
      "title": "わかったことを発表用にまとめる"
    },
    {
      "t": 3069.5,
      "title": "調べた結果の発表"
    }
  ]
};
