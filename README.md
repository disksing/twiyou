# twiyou

Twitter friend monitoring tool

----

twiyou（推油）是一款推特好友/推特数据监测工具。

![Screenshot 2022-10-07 145315](https://user-images.githubusercontent.com/12077877/194486031-dddb414b-905c-4422-9f47-ada4d2a39545.png)
![Screenshot 2022-10-07 145528](https://user-images.githubusercontent.com/12077877/194486415-3f53bc70-b82f-42ad-b7c8-f0051a5e7443.png)

## 搭建自己的部署

运行起来需要部署3个组件：数据库，数据抓取工具，监控面板，目前这3个组件都有免费的白嫖资源，此外需要有推特开发者账号。

- 数据库我用的是TiDB Cloud DevTier (https://tidbcloud.com/ )的免费资源，理论上兼容MySQL协议的应该都行，不过我没试。
- 数据抓取工具可以从 [Release](https://github.com/disksing/twiyou/releases) 下载，找台机器配置好环境变量定时运行程序就行（推荐5分钟运行一次）。也可以白嫖 GitHub Actions。
- 监控面板用的是Grafana，Grafana Cloud (https://grafana.com/ )也是可以创建免费的账号，添加数据库数据源之后，把监控模板导入就行。

## 搭建自己的部署（详细攻略）

### 1. 注册 twitter 开发者账号

注册地址：https://developer.twitter.com/en

申请成功后创建一个项目，然后把如图所示的Bearer Token记录下来备用。

![Screenshot 2022-10-07 152412](https://user-images.githubusercontent.com/12077877/194496392-21d8939d-0044-4070-b7a4-272963f5868d.png)

### 2. 注册 TiDB Cloud DevTier

注册地址：https://tidbcloud.com/

完成后创建一个免费的DevTier集群，然后在集群页面可以看到连接参数。

![Screenshot 2022-10-07 152706](https://user-images.githubusercontent.com/12077877/194497136-7d33c809-327d-4d63-9a01-ffc39f1e73f3.png)

### 3. 部署数据抓取工具

下载地址：https://github.com/disksing/twiyou/releases

然后配置好crontab和环境变量每5分钟运行一次就行，抓下来的数据都会写到数据库里面，本地不存数据。

环境变量设置如下图。

![Screenshot 2022-10-07 153706](https://user-images.githubusercontent.com/12077877/194498629-d8a8972f-2545-4469-b26e-c563b242f8b2.png)

注意DB_NAME需要提前自己连接数据库后创建，懒得手动创建的话直接填test数据库也行。

`TWITTER_USER_NAME`填自己的twitter id，当然你如果想监控别人的好友，填别人的id也可以。

如果自己没有服务器的话，可以使用免费的 GitHub Action 来跑：
1. fork 这个项目
2. 在 Repo Settings - Secrets 里配置好环境变量
3. 在 https://github.com/settings/tokens 里生成一个新的 Personal access token
4. 配置一个触发器定时发送类似下面的 HTTP 请求来触发 Github Action：
```shell
curl \
-X POST \
-H "Accept: application/vnd.github+json" \
-H "Authorization: Bearer <YOUR-GITHUB-TOKEN>" \
https://api.github.com/repos/<OWNER>/twiyou/actions/workflows/scrape.yml/dispatches \
-d '{"ref": "master"}'
```

这类第三方服务比较多，比如 Cloudflare worker，https://cron-job.org 都行。以 cron-job.org 为例：
<p align="center">
  <img width="45%" alt="image" src="https://user-images.githubusercontent.com/10510431/194699857-750ddb69-a377-4be0-a26d-9af24951c430.png">
  <img width="45%" alt="image" src="https://user-images.githubusercontent.com/10510431/194699890-629095d5-1619-41b7-a234-8fcbf3235358.png">
</p>

### 4. 配置 Grafana

注册地址：https://grafana.com

完成后在自己的Grafana实例添加MySQL数据源，参考截图，这里Timezone不要填。

![Screenshot 2022-10-07 154637](https://user-images.githubusercontent.com/12077877/194501010-32e40820-f282-4f4e-b627-392cd375ec33.png)

然后导入[预定义模板](https://raw.githubusercontent.com/disksing/twiyou/master/grafana/twitter-statistics.json)，选择之前创建的数据源。在第一轮数据抓完之前，可能不会立即看到数据。

## 一些问题和后续计划（请求出谋划策）

1. twitter开发者账号申请比较麻烦，而且API的rate limit等限制颇多，可以考虑换成绕过开发者API的方案。例如[twint](https://github.com/twintproject/twint)，不过我还没仔细研究。不好搞的话，或许可以考虑配置多个API key，轮换着用。

2. 想把自己的历史推文抓下来分析下，比如统计历史推文的传播表现，或者把发推历史跟follower数量变化做点关联分析之类，但是目前twitter API限制只能抓近一周的历史推文，所以这个功能可能依赖于上一个问题的解决。

## 排障 FAQ

1. `2022/10/13 07:37:39 Error 1130: Host '4.246.175.211' is blocked by traffic filter. See https://docs.pingcap.com/tidbcloud/connect-to-tidb-cluster#connect-via-standard-connection`

原因，TiDB Cloud 中 Cluster 未配置允许外部 IP 访问集群，需要 Create traffic filter，在其中增加 twiyou 二进制访问的出口 IP，如果没有固定 IP 可以允许任意 IP：`0.0.0.0/0`，只需要在 TiCloud 网页 Console 的集群下 Edit 就好了。

2. `client has multi-statement capability disabled. Run SET GLOBAL tidb_multi_statement_mode='ON' after you understand the security risk`

原因，报错是自解释的，不更改 twiyou/client 上的 multi-statement 模式的情况下，可以在 TiDB Cloud WebShell 里执行 `SET GLOBAL tidb_multi_statement_mode='ON'`，只需要在 TiCloud 网页 Console 的集群下点击 Connect 然后点击 WebShell，在 Shell 里输入密码登陆之后执行就好了。

