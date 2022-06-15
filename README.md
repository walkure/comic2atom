# comic2atom

Atomファイルを吐いてくれないWebコミックサイトをスクレイピングしてAtomを生成します。

## usage

適当なところにバイナリをおいて、`cron`とか`systemd.timer`で週に一回起動。

`comic2atom -targets https://site1/contents1,https://site1/contents2 -atom /var/www/atom`

## supported comic sites

公式でAtomとかRSSを吐いてくれればいいんですけどね…。

- 毎週水曜更新
  - [COMICメテオ](https://comic-meteor.jp/)
- 毎週金曜更新
  - [ストーリアダッシュ](https://storia.takeshobo.co.jp/)
  - [ガンマぷらす](https://gammaplus.takeshobo.co.jp/)
- ほぼ月～木
  - [コミックライド](https://www.comicride.jp/)
- 毎週火・金
  - [コミックヴァルキリー](https://www.comic-valkyrie.com/)
  
※ 自分が見たいとこだけ試したので、サイトで提供されてる全部のコンテンツで確実に動くわけではないです。

## author

walkure at 3pf.jp