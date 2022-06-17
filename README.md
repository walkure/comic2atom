# comic2atom

Atomファイルを吐いてくれないWebコミックサイト等をスクレイピングしてAtomを生成します。

## usage

適当なところにバイナリをおいて、`cron`とか`systemd.timer`で適当に起動。

`comic2atom -targets https://site1/contents1,https://site1/contents2 -list /foo/bar/list -atom /var/www/atom`

取得先URLは、`-targets`で書き連ねるのと`-list`でリストファイル(1URI毎に1行)を渡すのと両方対応(片方だけでも良い)しています。

## supported sites

公式でtopic単位のAtomとかRSSを吐いてくれればいいんですけどね…。

- 毎週水曜更新
  - [COMICメテオ](https://comic-meteor.jp/)
- 毎週金曜更新
  - [ストーリアダッシュ](https://storia.takeshobo.co.jp/)
  - [ガンマぷらす](https://gammaplus.takeshobo.co.jp/)
- ほぼ月～木
  - [コミックライド](https://www.comicride.jp/)
- 毎週火・金
  - [コミックヴァルキリー](https://www.comic-valkyrie.com/)
- 随時
   - [小説家になろう](https://syosetu.com/)

※ 自分が見たいとこだけ試したので、サイトで提供されてる全部のコンテンツで確実に動くわけではないです。

## author

walkure at 3pf.jp