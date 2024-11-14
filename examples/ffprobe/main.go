package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffprobe"
)

func main() {
	ctx := context.Background()

	fp, err := ffprobe.New(
		ffprobe.WithDebug(true),
		ffprobe.WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36 Edg/130.0.0.0"),
	).
		Input("http://upos-sz-estgoss.bilivideo.com/upgcxcode/05/64/26733186405/26733186405-1-192.mp4?e=ig8euxZM2rNcNbRV7wdVhwdlhWdMhwdVhoNvNC8BqJIzNbfq9rVEuxTEnE8L5F6VnEsSTx0vkX8fqJeYTj_lta53NCM=&uipk=5&nbs=1&deadline=1731563657&gen=playurlv2&os=upos&oi=1866713016&trid=03320e0909fa4d6491c858f81cf3a21bO&mid=0&platform=html5&og=08&upsig=9caf64d03b6e3b7d88065fd20cf0e2f1&uparams=e,uipk,nbs,deadline,gen,os,oi,trid,mid,platform,og&bvc=vod&nettype=1&orderid=0,3&buvid=&build=7330300&f=O_0_0&bw=104752&logo=80000000").
		Extract(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	vStream := fp.GetFirstVideoStream()
	if vStream == nil {
		fmt.Println("file has no video stream")
		return
	}

	fmt.Println(vStream)
}
