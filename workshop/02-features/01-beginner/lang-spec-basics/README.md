# プログラミング言語仕様を読もう

あなたは超久々に Go コーディングをします。  
言語仕様についてド忘れしていることもあるので、Go 言語仕様を読んでみましょう。

## 設問 1: for文

カウンター変数を使ったものは当然あるとして `in` のような構文があったはずですが、Go ではどう書くのでしょうか？

<details>
<summary>ヒント</summary>

- 特に無し

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://go.dev/ref/spec を開き、目次から「For statements」を選ぶ。
2. `for` 文の構文を確認する。

**答え**

- 以下のように、`for` 文は 3 種類の構文を持っています。
- このうち、`in` のような構文はありませんが、 `range` キーワードを使った構文があり、これが `in` のような意味合いを持っています。

```
ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
```

[do it](https://go.dev/play/p/WCZlhaPrgG_p)

</details>

---

## 設問 2: switch文

Go の switch 文は、C 言語の switch 文と違って、break が不要なことは知っています。  
では、C 言語の switch 文のように、1 つの case にマッチしたら、次の case も実行されるようにするにはどうすればよいでしょうか？

<details>
<summary>ヒント</summary>

- 特に無し

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://go.dev/ref/spec を開き、目次から「Switch statements」を選ぶ。
2. `fallthrough` キーワードの説明を確認する。

**答え**

- `fallthrough` キーワードを使うことで、C 言語の switch 文のように、1 つの case にマッチしたら、次の case も実行されるようにできます。

```go
switch x {
case 1:
    fmt.Println("one")
    fallthrough
case 2:
    fmt.Println("two")
}
```

[just do it](https://go.dev/play/p/HumLQAbZ-eQ)


</details>

---

<details>
<summary>こぼれ話: </summary>

[https://go.dev/ref/spec](https://go.dev/ref/spec)（Go言語仕様書）は、プログラミング言語の仕様書としては異例なほど短く、1ページのHTMLにまとまっているのが最大の特徴です。  
C++やJavaの仕様書が分厚い辞書のようであるのに対し、Goの仕様書は「休日の午後だけで読み切れる」ことを意図して設計されています。

</details>

---

## 調査の入り口

- https://go.dev/ref/spec
