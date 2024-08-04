# comic2atom

Atomファイルを吐いてくれないWebコミックサイト等をスクレイピング等してAtomを生成します。

## usage

### converter

適当なところにバイナリをおいて、`cron`とか`systemd.timer`で適当に起動。

`comic2atom -targets https://site1/contents1,https://site1/contents2 -list /foo/bar/list -atom /var/www/atom`

取得先URLは、`-targets`で書き連ねるのと`-list`でリストファイル(1URI毎に1行)を渡すのと両方対応(片方だけでも良い)しています。

### proxy

RSSリーダから到達できる適当なところで起動しておき、RSSリーダに登録するURIのprefixに当該proxyのURIをつける。

e.g. `http://localhost:18080/entry/https://www.example.com/comic/1`

### Docker

`docker run --rm -it --mount type=bind,source=/path/to/output,target=/output ghcr.io/walkure/comic2atom/converter:latest -targets "https://site1/contents1,https://site1/contents2" -atom /data/`

`docker run --rm -it -p 18080:8080 ghcr.io/walkure/comic2atom/proxy:latest`


## supported sites

公式でtopic単位のAtomとかRSSを吐いてくれればいいんですけどね…。

- 毎週水曜更新
  - [COMICメテオ](https://comic-meteor.jp/)
- 毎週金曜更新
  - [ストーリアダッシュ](https://storia.takeshobo.co.jp/)
  - [ガンマぷらす](https://gammaplus.takeshobo.co.jp/)
- 毎週火・金
  - [コミックヴァルキリー](https://www.comic-valkyrie.com/)
- 随時
  - [小説家になろう](https://syosetu.com/)
  - [カクヨム](https://kakuyomu.jp/)
  - [COMIC FUZ](https://comic-fuz.com/)
  - [カドコミ](https://comic-walker.com/)
  - [コミックNewtype](https://comic.webnewtype.com/)

※ 自分が見たいとこだけ試したので、サイトで提供されてる全部のコンテンツで確実に動くわけではないです。

## author

walkure at 3pf.jp
