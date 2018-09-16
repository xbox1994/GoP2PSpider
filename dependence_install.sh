#!/usr/bin/env bash
sudo add-apt-repository ppa:gophers/archive
sudo apt-get update
sudo apt-get install golang-1.10-go -y
echo "PATH=/usr/lib/go-1.10/bin:$PATH" >> ~/.profile
source ~/.profile

sudo apt-get install openjdk-8-jre -y
curl -L -O https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-5.6.11.deb
sudo dpkg -i elasticsearch-5.6.11.deb
sudo /etc/init.d/elasticsearch start

cd ~
mkdir go
cd go
mkdir src
cd src
git clone https://github.com/xbox1994/GoP2PSpider.git
cd GoP2PSpider

go get gopkg.in/olivere/elastic.v5
go get github.com/xbox1994/bencode
go get golang.org/x/time/rate