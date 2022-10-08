# SPI接続のE-Paperを制御するプログラム

## 概要
InflaxDBにため込んだ情報をgrafanaで参照するのがめんどくさいので、一日の最高と最低、平均、1時間の平均が計算して出力するプログラム

## 前準備
SPIを使用しているので有効にする

本語フォントファイルとして、IPAex を使用するのでインストールする
```
sudo apt install fonts-ipaexfont -y
```


## 読み取り方法

以下の通りにすると11:54移行でco2の情報をすべて取得
curl http://192.168.0.6:8086/query?db=senser --data-urlencode "q=SELECT * FROM senser_data WHERE time >= '2022-10-07T02:54:00Z' AND type='co2'"|jq


## 自動起動について
epaperifdb.service ファイルを/etc/systemd/system/に置くことでsystemctlコマンドで制御可能実行用のスクリプトは対象場所におく


## 閾値について

積み重ねの値で、99パーセントを超えるものは取得値から排除
比較値の計算式 X=現在の値
四捨五入(絶対値(log10((X-最小値)/(平均値-最小値+1)))*10)