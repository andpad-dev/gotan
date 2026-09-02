# add-scenario

Go探無比のシナリオを、調査体験、一次情報、読みやすさまで含めて作成するskillです。

作成時から次を確認します。

- 文脈を調査対象の性質だけを保つ中立表現へ置換し、不要なら削る
- 意味を調べない値には、題材固有の意味を持たない中立値を使う
- 1文60文字以内、漢字比率30〜40%、STR15〜30%を目標にする
- 観測・制約・次の操作を見出し、箇条書き、太字で拾いやすくする
- 図が文章量や読み戻しを減らす場合だけMermaidを使う
- Go解析器で最長文、漢字比率、STR、MDD推定値、n-gram平均サプライザル、CRS推定値を計測する
- PR前に、作者と履歴を共有しない独立コンテキストで全許可ロールを実プレイする
- 改善を反映した後、最終内容を再プレイしてからDraft PRを作る

MDDはKagomeの品詞を使う前方係り受けの推定値です。平均サプライザルは、
対象外のworkshop READMEで学習したadd-alpha単語bigramの正規化済み条件付きサプライザルです。
どちらもneural LMやbase ONNXモデルの値とは呼びません。これらから計算する`crs_estimate`も、
モデル由来のCRSと同一比較しません。

```bash
go -C .agents/skills/playtest-workshop-feedback/scripts test ./...
go -C .agents/skills/playtest-workshop-feedback/scripts run . \
  ../../../../workshop/<category>/<difficulty>/<topic>/README.md
```
