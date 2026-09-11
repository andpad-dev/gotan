# ワークショップ進行ガイド・シナリオ一覧

当日は、この README を進行ガイドとして使います。
ファシリテーターはこのページを画面共有し、上から順に沿って進行します。
参加者のみなさんも、このページを手元で開いて進行を追ってください。

事前準備（班ごとの Issue 作成など）、当日のスタッフの動き、チュートリアルのデモ台本は [STAFF.md](STAFF.md) にまとめています。
班で参加する人は [チームでの進め方](TEAM_GUIDE.md) も開いてください。

## 最初に開くページ

- [このページ](README.md): 当日の進行と全シナリオの一覧
- [チームでの進め方](TEAM_GUIDE.md): ペアの分かれ方と、Issue への記録方法
- [Go Playground](https://go.dev/play/): ローカルに Go がなくても、小さな `package main` をブラウザで実行できる

当日のやりとりは、班に割り当てられた **Issue** に集約します。調べたことも、共有したい URL も、そこへコメントしてください。
[Discord](https://discord.gg/xBrQdb3ehW) も用意していますが、当日必須ではありません。ワークショップが終わったあとも残るので、後日の質問はこちらへどうぞ。

Go Playground ではコードを貼って **Run** を押すと実行できます。班へコードを渡すときは **Share** を押し、生成された
`https://go.dev/play/p/...` の URL を Issue に貼ってください。受け取った人は同じソースを開けます。版依存のシナリオでは、
サンプルが表示する Go のバージョン、または画面に表示されたバージョンも一緒に記録します。

シナリオによっては、手元の Go で動かす手順があります。その場合は同梱の `go.mod` が必要な Go のバージョンを指定しています。
手元の Go がシナリオの要求より古い場合、`go run` はツールチェーンのダウンロードを始めます。回線が細いときは、その時間も見込んでください。

用語は次のように使い分けます。

- **シナリオ**: 1 枚の README にまとまった教材全体
- **設問**: シナリオ内の `設問 1`、`設問 2` などの番号付きの問い
- **テーマ**: 班が当日調べる対象。用意されたシナリオでも、持ち込んだ疑問でもよい

### ヒントの使い方

- **時間制のワークショップ**: ヒントも答えも開いて構いません。ヒントを使ったら、示された検索語やページを自分でたどります。探索ワークの残り20分でファシリテーターから声をかけます。
- **時間を決めない自習**: まずヒントを閉じて試し、5 分進展がない、または異なる調査を 2 回試して進めないときに開きます。

答えの結論だけを読むのではなく、**調査ルートを再現して、自分たちの Issue に操作と発見を残す**ことを完了条件にします。

## シナリオ一覧

すべてのシナリオは [シナリオ一覧](SCENARIOS.md) にまとめています。テーマや難易度から選ぶときは、こちらを開いてください。

## タイムテーブル（予定: 90分）

| 時間 | 内容 |
|:---|:---|
| 00:00 - 00:05 | イントロダクション、趣旨説明 |
| 00:05 - 00:20 | チュートリアルワーク（用意されたシナリオで調べ方の練習） |
| 00:20 - 00:25 | グループ分け |
| 00:25 - 00:30 | アイスブレイク（3分固定）とテーマ選択 |
| 00:30 - 01:15 | 探索ワーク（45分） |
| 01:15 - 01:30 | 全体での共有、クロージング |

---

## 00:00 イントロダクション（5分）

### このワークショップの目的

みなさん、こんなお悩み、ありませんか？

- 同僚やLLMの書いたGoのコードがよく分からないので調査したい
- Goに関する質問をLLMにしたが、LLMの回答が正しいのかどうか判断が付かない
- 他の言語にはあるのにGoにはない、もしくはGo特有の機能が何故あるのか、議論の背景を深堀りしたい
- 次回以降のGo Conferenceへ質の高いProposalを出して登壇を目指したい
- Goのコントリビューターを目指したい！

検索結果、LLM、技術ブログは調査の入口になりますが、回答が現在の仕様と一致するとは限りません。仕様の背景や採用されなかった案まで確かめるには、要約の根拠となった資料をたどる必要があります。

Go の公式ドキュメント、リリースノート、Issue、ソースコードを順にたどり、手元の実行結果と照合します。調査の結論に加え、Go チームが選んだ設計とその条件を自分で説明できる状態を目指します。

### 今日の流れ

1. **チュートリアルワーク:** ファシリテーターの説明を手元で追いながら、公式ドキュメントをたどる「調べ方のコツ」を体験します。
2. **グループ分け:** 挑戦したい難易度ごとに、4名程度のグループに分かれます。
3. **テーマ選択:** グループごとに **「今日深く調べたいテーマ」** を決め、班に割り当てられた Issue に宣言します。
   用意されたシナリオから興味のあるものを選んでもよいですし、日々の開発で気になっている自分たちの疑問をテーマにしても構いません。
4. **探索ワーク:** 班を2人ずつのペアに分け、設問を分担して一次情報をたどります。調べた過程は用意された Issue に「調査ログ」として記録してもらいます。
5. **全体共有:** いくつかの班をピックアップして、「どうやって調べたか」「どんな発見があったか」を全体に発表してもらいます。

「正解」を知ることよりも、**「どのようにしてその情報へ辿り着くか」という調査の過程** を大切にしてください。
ヒントと答えは各シナリオの README に折りたたんで最初から公開しています。
行き詰まったら自由に開いて構いません。

- **設問を全部解き切る必要はありません。**
- リポジトリは公開したままにしておくので、後日取り組んでも構いません。

### LLM の使い方

調べたいことをそのまま LLM に聞いて、返ってきた答えで済ませるのは避けてください。
今日持ち帰ってほしいのは結論ではなく、一次情報にたどり着くまでの道筋です。答えだけを受け取ると、その道筋を通らないまま終わってしまいます。

使ってはいけないわけではありません。次のような使い方は、一次情報へ向かう助けになります。

- **英語の一次情報を翻訳する、要約する。** 一次情報を開いたうえで意味を取るための翻訳は歓迎します
- どのドキュメントの、どのあたりに書かれていそうかの見当をつける
- 英語のドキュメントを読んだあと、自分の理解が合っているかを確かめる
- LLM の回答を出発点にして、その根拠を一次情報で裏取りする

班の Issue に残す調査ログには、LLM に聞いた内容ではなく、実際に開いたページと、そこで確かめた事実を書いてください。

---

## 00:05 チュートリアルワーク（15分）

[チュートリアル](00-tutorial/README.md) を題材に、調べ方をスライドで追体験します。

ファシリテーターがスライドを送りながら手順を説明するので、参加者は次の2つを手元で開き、同じ操作を追ってください。

- [チュートリアルの README](00-tutorial/README.md)
- <https://pkg.go.dev/fmt>

デモは次の流れで進みます。

1. シナリオのコード `fmt.Printf("%#[1]v %[1]T\n", value)` を [Go Playground](https://go.dev/play/p/RTNSvn_p2Ai) で実行し、出力を確認する。
   **Share** を押して URL で同じコードを共有できることも確認する。
2. 設問1: Ctrl+F で本文を検索し、`%v`・`%T`・`#` の意味をドキュメントから突き止める。
3. 設問2: `[1]` の正体を「Explicit argument indexes」セクションから突き止める。

20 分のチュートリアルで、ワークショップ全体に使う **調査の型** を練習します。

1. 公式ドキュメントを開く。
2. キーワードで該当セクションに絞る。
3. 手を動かして挙動を確かめる。

カテゴリ別の逆引きは、各カテゴリの README（例: [01-packages の調べ方](01-packages/README.md)）にまとまっています。
探索ワークで困ったら、まずそこに戻ってください。

また、各シナリオの README には、ヒントと答えが `<details>`（折りたたみ）で最初から公開されています。
答えには **調査ルート**（どの一次情報をどの順で辿るか）が書かれているので、答えを開いた場合でも、そのルートを自分でなぞり直してみましょう。

---

## 00:20 グループ分け（5分）

参加者のレベルに合わせ、次の3段階のシナリオを用意しています。

- 初級 (Beginner): 「◯◯を見かけました。どんなものか調べてみましょう」
- 中級 (Intermediate): 「◯◯をやりたいです。どういうふうにやればいいか調べよう」
- 上級 (Advanced): 「なんでこうなってるの？背景を調べよう」

自分が挑戦してみたいレベルごとに集まり、4名程度の班を作ってください。
スタッフが挙手で人数を確認し、エリアへ誘導します。

**レベル選びの目安**

- 上級者が中級・初級のシナリオを解くのはまったく問題ありません。
- 逆に、初級者がいきなり上級のシナリオに挑むと時間内に終わらない可能性が高いので、背伸びするなら中級までがおすすめです。
- 自分のレベルが分からない場合、Go歴がある程度長いならちょっと上を目指してみても良いかもしれません。

班が決まったら、**班専用の Issue の URL** を確認してください。
班専用の Issue を、今日の調査ログと発表資料を兼ねるノートとして使います。

---

## 00:25 アイスブレイクとテーマ選択（5分）

まず **3分** とって、班内で自己紹介をしましょう（名前・Go歴・今日調べてみたいことを一言ずつ）。

続いて、[チームでの進め方](TEAM_GUIDE.md) を開き、班ごとに **「今日深く調べたいテーマ」** を1つ決めます。
テーマの決め方は次の2通り、どちらでも構いません。

1. **用意されたシナリオから選ぶ:** [シナリオ一覧](SCENARIOS.md) から、興味のあるものを選びます。カテゴリは次の4つです。
   - [01-packages](01-packages/): Goのパッケージに関する出題
   - [02-features](02-features/): Goの機能に関する出題
   - [03-cmd-tools](03-cmd-tools/): Goのtoolに関する出題
   - [04-deep-dive](04-deep-dive/): Goの歴史などGoを更に深く知りたい人向けの出題
2. **自分たちの疑問を持ち込む:** 日々の開発で気になっていた疑問を、そのままテーマにします。

**テーマ選択のコツ**

- 探索ワークは 45 分です。「1つの疑問に絞る」くらいの粒度がちょうどよいです。
- 持ち込みテーマの場合は、「なぜ◯◯はこういう仕様なのか」「◯◯はどう実装されているのか」のように、一次情報にあたって答えられる形の問いに言い換えてみましょう。

テーマが決まったら、班の Issue に誰か1人がテーマをコメントしてください。
ペアの分け方は探索開始後の 3 分で決めます。

---

## 00:30 探索ワーク（45分）

設定したテーマについて、班で調査を進めます。

**全設問を解き切ることは目的ではありません。** 1問を確実に深掘りするほうが、今日の目的に合っています。
解けなかった分は持ち帰って続けられます。

### 進め方

1. シナリオは**班全体で1つ**に取り組みます。まず最初の3分で、シナリオの **調査の入り口** を開き、どこから調べ始めるかを班で決めます。カテゴリの README（逆引き手順）を先に見ても構いません。
2. **設問はペアで分担します。** 4人なら2人ずつの2組に分かれ、どの設問を担当するかを決めてください。
3. 公式ドキュメントやソースコードを辿りながら、分かったことを確かめていきます。挙動が気になったら Go Playground や手元の `go run` で実際に動かしてみましょう。
4. 調査の過程は、**見つけた人がその場で** 班の Issue にコメントで残します。記録の担当は決めず、全員で書き込んでいきましょう。
5. 終盤にペアどうしで持ち寄り、互いの発見を一つのコメントとしてまとめます。

### 調査ログの書き方

きれいにまとめる必要はありません。
URL だけでは後から調査を再現できません。次の項目をそのつどコメントしてください。

```text
- 担当した問い: <何を確かめようとしたか>
- 見た場所: <URLと該当セクション名>
- 操作・検索語: <どうやってそこへ到達したか>
- 分かったこと: <根拠から言えること>
- 次の疑問・行き詰まり: <あれば>
```

残したログが、最後の全体共有でそのまま発表資料になります。後日ブログを書くときの下書きにもなります。

### 行き詰まったら

まず、設問の前にある **調査の入り口** に戻ってください。まだ開いていない資料が残っているはずです。順番に見るものではないので、気になるものから入り直して構いません。

それでも進まないときは、次を試します。

- カテゴリの README（逆引き手順）に戻って、別の切り口を探す。
- **ヒントを開く。** 最初から開いて構いません。検索語や一次情報の入口として使ってください。
- **答えを開く。** 残り20分でファシリテーターから声をかけます。結論をコピーするのではなく、答えに書かれた **調査ルート** を自分でなぞり直すのがおすすめです。
- スタッフが会場を巡回しています。遠慮なく声をかけてください。
- 記述どおりに動かない、設問の意図が読み取れない、といったシナリオ側の問題に見えるときは、[シナリオの不備報告](https://github.com/andpad-dev/gotan/issues/new?template=scenario-issue.md)から Issue を立ててください。気づいたことだけ書いて、調査は止めずに先へ進んでください。

### 終盤の合図

ファシリテーターが2回声をかけます。

- **残り20分:** ここから答えを開いて構いません。時間内に手ぶらで終わらないための区切りです。
- **残り10分:** 班の Issue に **発表用まとめ** のコメントを書き始めてください。

```text
## 発表まとめ
- 調べたテーマ:
- 最初の観測:
- 調査ルート: <出発点 → 検索語/リンク → 一次情報 → 実測>
- 分かったこと:
- まだ分からないこと:
- 一番の発見:
```

---

## 01:15 全体での共有（15分）

**3班ほどをピックアップして発表してもらいます。** 初級・中級・上級から拾います。全班は回りません。
ファシリテーターが班の Issue を画面に映すので、PC を繋ぎ替える必要はありません。

発表では次の3点を話してください。1班2分が目安です。

1. **調べたテーマ:** 何を調べようとしたか。
2. **調査ルート:** どの一次情報を、どの順で辿ったか。
3. **一番の発見:** 意外だったこと、面白かったこと。「ここで行き詰まった」という共有も立派な発見です。

結論の正しさよりも、「どう辿り着いたか」を中心に話すのがこのワークショップ流の発表です。

発表しなかった班の調査ログも、すべて Issue に残っています。気になるテーマがあれば、あとから読んでみてください。

### クロージング

最後に、今日の持ち帰りを3つ確認して締めくくります。

- 今日体験した調査の型はシンプルです。「公式ドキュメントを開く → キーワードで絞る → 手を動かして確かめる」。明日からの開発でも、ググる前・LLMに聞いた後に、ぜひ一次情報を開いてみてください。
- このリポジトリは公開されています。今日解かなかったシナリオも、家で同じように解けます。
- 一次情報を辿った先には、Go Conference への Proposal や Go 本体へのコントリビュートという道もあります。今日の調査ログが、その第一歩です。

### 続きは Discord で

冒頭で開いてもらった [Discord](https://discord.gg/xBrQdb3ehW) は、今日で閉じません。解ききれなかったシナリオの相談も、後日の質問も、こちらでどうぞ。
シナリオの不備に後から気づいたときも、[Issue](https://github.com/andpad-dev/gotan/issues/new?template=scenario-issue.md) で教えてください。

今日 Issue に残した調査ログは、そのままブログの下書きになります。
どこから調べ始めて、何で行き詰まって、どう抜けたか。その道筋は調べた本人にしか書けません。
結論は調べれば誰でも同じところに着きますが、どう着いたかはあなただけのものです。書いたら Discord で教えてください。X に `#gocon26` を付けてポストしてもらえると、当日いなかった人にも届きます。

シナリオの設問におかしなところがあれば、[Issue](https://github.com/andpad-dev/gotan/issues) で教えてください。当日その場で気づいた違和感が、次の開催の改善になります。

お疲れさまでした！


----------------------


# Workshop Progress Guide & Scenario Index

Use this README as the workshop guide on the day. The facilitator shares this page and proceeds from top to bottom. Participants should open it locally and follow along.

Preparation, staff activities, and the tutorial demo script are in [STAFF_en.md](STAFF_en.md). Teams should also open [How to work as a team](TEAM_GUIDE_en.md).

## Pages to Open First

- [This page](README.md): Workshop flow and scenario index
- [How to work as a team](TEAM_GUIDE_en.md): Icebreaker, roles, timing, and issue notes
- [Go Playground](https://go.dev/play/): Run small Go programs without a local installation
- [Discord](https://discord.gg/xBrQdb3ehW): Workshop communication and team discussion

Open Discord first so teams can share Playground URLs, record blockers, and post their issue URLs. The channel remains available after the workshop for later questions.

In Go Playground, paste the code and press **Run**. Press **Share** to create a `https://go.dev/play/p/...` URL and put it in the team issue. For version-dependent scenarios, record the Go version shown by the sample or Playground.

Some scenarios require a local Go installation. Their `go.mod` files specify the required version. If the installed Go version is older, `go run` may download a toolchain.

Use these terms consistently:

- **Scenario**: All learning material collected in one README.
- **Question**: A numbered question such as `Question 1`.
- **Theme**: What the team investigates, either a provided scenario or its own question.

### Using Hints

- **Time-boxed workshop**: Hints may be opened from the beginning. Follow the suggested search terms and pages yourself.
- **Self-paced study**: Try first with hints closed; open them after five minutes without progress or two unsuccessful approaches.

Answers are also allowed. Do not only read the conclusion: reproduce the investigation path and record the operations and discoveries in the team issue.

## Scenario Index

All scenarios are listed in [SCENARIOS_en.md](SCENARIOS_en.md). Choose from it by theme and difficulty.

## Timetable (Planned: 90 minutes)

| Time | Content |
|:---|:---|
| 00:00 - 00:05 | Introduction and purpose |
| 00:05 - 00:20 | Tutorial work |
| 00:20 - 00:25 | Team formation |
| 00:25 - 00:30 | Icebreaker (fixed at 3 minutes) and theme selection |
| 00:30 - 01:15 | Exploration work (45 minutes) |
| 01:15 - 01:30 | Whole-group sharing and closing |

---

## 00:00 Introduction (5 minutes)

### Purpose of the workshop

This workshop is for investigating Go code written by colleagues or an LLM, checking whether an LLM's answer is correct, understanding why Go has or does not have a feature, preparing a high-quality Proposal for a future Go Conference, and learning how to contribute to Go.

Search results, LLMs, and technical blogs are useful starting points, but their answers do not necessarily match the current specification. To understand the background of a design or proposals that were not adopted, follow the sources behind the summary.

Trace Go's official documentation, release notes, Issues, and source code in order, then compare them with local execution results. The goal is to explain not only the conclusion but also the design Go chose and the conditions around it.

### Today's flow

1. **Tutorial work:** Follow the facilitator's live demo and practice tracing primary documentation.
2. **Team formation:** Form groups of about four people according to the difficulty you want to try.
3. **Theme selection:** Choose one theme to investigate deeply and announce it in the team's Issue. You may choose a prepared scenario or bring a question from daily development.
4. **Exploration work:** Divide into pairs, split the questions, and trace primary sources. Record the investigation in the team's Issue.
5. **Whole-group sharing:** Several teams present how they investigated and what they discovered.

The investigation process matters more than simply knowing the answer. Hints and answers are available in collapsible sections from the beginning; open them whenever you are stuck. You do not need to finish every question, and you may continue with the repository after the workshop.

### Using LLMs

Do not ask an LLM a question and stop at its answer. The goal today is to learn the path to primary information. LLMs are welcome as support for translating or summarizing English primary sources after opening them, locating likely documents and sections, checking your understanding after reading, and finding a starting point whose evidence you then verify yourself.

In the team's investigation log, record the pages you actually opened and the facts you confirmed, not only what an LLM said.

---

## 00:05 Tutorial work (15 minutes)

Use the [tutorial](00-tutorial/README.md) to practice the investigation method in a live demo. Open the [tutorial README](00-tutorial/README.md) and <https://pkg.go.dev/fmt> locally and follow the facilitator.

1. Run `fmt.Printf("%#[1]v %[1]T\n", value)` in the [Go Playground](https://go.dev/play/p/RTNSvn_p2Ai) and confirm the output. Press **Share** to verify that the same code can be shared by URL.
2. For Question 1, use page search to determine the meanings of `%v`, `%T`, and `#`.
3. For Question 2, identify `[1]` in the “Explicit argument indexes” section.

Practice the workshop's investigation pattern: open the official documentation, narrow it with keywords, and run a small experiment. Category reverse-lookup guidance is in each category README, and every scenario includes collapsible hints and answers with an investigation path.

---

## 00:20 Team formation (5 minutes)

The scenarios have three levels:

- **Beginner:** “I saw something. Let us find out what it is.”
- **Intermediate:** “I want to do something. Let us find out how.”
- **Advanced:** “Why does it work this way? Let us investigate the background.”

Form groups of about four people at the level you want to try. Advanced participants may solve easier scenarios, while beginners should generally start no higher than intermediate. Once your group is formed, confirm the URL of its dedicated Issue and use it as both the investigation log and presentation notes.

---

## 00:25 Icebreaker and theme selection (5 minutes)

Take **three minutes** for introductions: name, Go experience, and one thing you want to investigate today. Then open [How to work as a team](TEAM_GUIDE.md) and choose one theme. Select a prepared scenario from [SCENARIOS.md](SCENARIOS.md), or bring a question from daily development.

Keep the scope to roughly one question for the 45-minute exploration. Rephrase a brought-in topic as something that can be answered from primary information, such as “Why does this specification behave this way?” or “How is this implemented?” Have one person comment the theme in the team Issue. Decide the pair split during the first three minutes of exploration.

---

## 00:30 Exploration work (45 minutes)

The whole team works on one scenario. During the first three minutes, open its **investigation entry points** and decide where to start; the category README can also be used first. Split the questions between pairs, follow official documentation and source code, and run a Playground or local `go run` experiment when behavior is unclear.

Record discoveries in the team Issue as they happen:

```text
- Question: <what you tried to verify>
- Source: <URL and section>
- Operation/search term: <how you got there>
- Finding: <what the evidence shows>
- Next question/blocker: <if any>
```

If you get stuck, return to the entry points, try another route in the category README, open the hint, or open the answer and retrace its **investigation path** instead of copying its conclusion. Ask staff for help, and report apparent scenario defects through the [scenario issue form](https://github.com/andpad-dev/gotan/issues/new?template=scenario-issue.md) without stopping the investigation.

At **20 minutes remaining**, answers may be opened. At **10 minutes remaining**, begin a presentation summary in the team Issue:

```text
## Presentation summary
- Theme investigated:
- First observation:
- Investigation path: <starting point -> search term/link -> primary source -> measurement>
- Findings:
- Open questions:
- Most important discovery:
```

---

## 01:15 Whole-group sharing (15 minutes)

About three teams will present, with examples from the beginner, intermediate, and advanced levels. Each team has about two minutes:

1. What theme did you investigate?
2. Which primary sources did you follow, and in what order?
3. What was the most surprising or interesting discovery? Sharing where you got stuck is also valuable.

The method matters more than presenting a polished conclusion. All teams' logs remain in their Issues for later reading.

### Closing

Take away three points: open official documentation, narrow it with keywords, and verify by running code; the repository remains available for unfinished scenarios; and following primary sources can lead to a Go Conference Proposal or contribution to Go itself.

The [Discord](https://discord.gg/xBrQdb3ehW) remains open for questions and unfinished scenarios. Report later-discovered scenario problems through the [Issue form](https://github.com/andpad-dev/gotan/issues/new?template=scenario-issue.md). Your investigation log can become a blog draft, because the path you took is uniquely yours. You can also share it on X with `#gocon26`.

