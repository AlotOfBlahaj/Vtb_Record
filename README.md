# Vtb_Record Vtuber直播监控-录播|QQ提醒

[![Maintainability](https://api.codeclimate.com/v1/badges/de4fc066a73b9822e6c5/maintainability)](https://codeclimate.com/github/fzxiao233/Vtb_Record/maintainability) ![commit activity](https://img.shields.io/github/commit-activity/m/fzxiao233/Vtb_Record?style=flat-square)

## 介绍

这是由[live_monitor系列](https://github.com/fzxiao233/live_monitor_server)使用Go重写后的新全家桶

功能:

- QQ提醒
- 视频下载
- 视频上传
- 同传记录

目前支援如下平台的监控:

- Youtube

- Twitcasting

- Bilibili

## 特性

- GO卓越的并发监控

- 超低内存和CPU消耗(在闲置状态下平均占用20m内存，检测时占用视监控人数定)

- 简单易用，只需配置好config后启动即可

- 跨平台，支持Windows/Linux/MacOS

## 使用方式

一、 在[Release](https://github.com/fzxiao233/Vtb_Record/releases)下载对应平台已编译好的二进制文件

二、 安装依赖streamlink方法见[文档](https://streamlink.github.io/install.html)

三、 配置config.json文件

- 拷贝config_example.json并改名config.json

- 编辑config.json以设置功能

```jsonc
{
  "EnableProxy": false,  // 是否启用代理
  "Proxy": "127.0.0.1:10800",  //代理地址，应为socks5代理
  "CriticalCheckSec": 30, // 检测间隔
  "NormalCheckSec": 30, // 检测间隔
  "DownloadQuality": "best",  // 配置下载画质 best为最佳画质 建议不调整 可选: best 1080p60 720p
  "DownloadDir": "/home/ubuntu/Matsuri",  // 下载目录 注意后无斜杠
  "EnableTS2MP4": true,  // 是否启用ts转码mp4（关闭后断流文件不会合并
  "Module": [
    {
      "Name": "Youtube",  // 模块名，以下类推
      "Enable": true,  // 是否启用该模块
      "Users": [  // 监测对象配置，详细见下文
        {
          "TargetId": "UCQ0UDLQCjY0rmuxCDE38FGg",
          "Name": "natsuiromatsuri",
          "NeedDownload": true,
          "NeedCQBot": true,
          "CQHost": "",
          "CQToken": "",
          "QQGroupID": [
            ""
          ],
          "TransBiliId": "336731767"
        }
      ]
    },
    {
      "Name": "Twitcasting",
      "Enable": true,
      "Users": [
        {
          "TargetId": "natsuiromatsuri",
          "Name": "natsuiromatsuri",
          "NeedDownload": true,
          "NeedCQBot": true,
          "CQHost": "",
          "CQToken": "",
          "QQGroupID": [],
          "TransBiliId": "336731767"
        }
      ]
    },
    {
      "Name": "Bilibili",
      "Enable": true,
      "Users": [
        {
          "TargetId": "336731767",
          "Name": "natsuiromatsuri",
          "NeedDownload": false,
          "NeedCQBot": true,
          "CQHost": "",
          "CQToken": "",
          "QQGroupID": [
          ]
        }
      ]
    }
  ]
}
```

- Users检测对象配置详解

```jsonc
{
    "TargetId": "UCQ0UDLQCjY0rmuxCDE38FGg", // ①见下文
    "Name": "natsuiromatsuri", // 对象名称，用以设置下载目录以及和上传端交互
    "NeedDownload": true,  // 是否启用下载
    "NeedCQBot": true,  // ②是否启用QQ机器人bot通知，配置见下文
    "CQHost": "",  
    "CQToken": "",
    "QQGroupID": [
    ""
    ],
    "TransBiliId": ""  // ③用以捕获在B站弹幕中的同传，配置见下文
}
```

① TargetId配置

均为粗体部分

- Youtube, 如 www.youtube.com/channel/ **UCp6993wxpyDPHUpavwDFqgg** 

- Twitcasting, 如 twitcasting.tv/ **natsuiromatsuri**

- Bilibili, 如 space.bilibili.com/ **336731767**

② QQ机器人的配置

- 后端仅支持Coolqq， 详情见对应文档，需求安装HTTPAPI

- CQHost为HTTPAPI的地址

- CQToken为在HTTPAPI中设置的token（如为设置为空

- QQGroupID为需要发送的群号，各群号间用,分割

③ TransBiliId设置

- TransBiliID为该监控目标对应的Bilibili用户编号（粗体部分）space.bilibili.com/ **336731767**

- 同传指在直播间发送的带有【】标记的内容，会保存为txt文本

- 使用本功能需配置后端[bilibili-danmaku-translation-recorder](https://github.com/fzxiao233/bilibili-danmaku-translation-recorder) 

四、启用本程序

在命令行下执行即可

## 支持

如对本程序使用遇到问题或有建议和意见请发送issue或发送邮件至fzxiao@dd.center
