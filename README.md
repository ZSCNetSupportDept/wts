# 中山学院网维报修系统
> 第三版


## 介绍
用于电子科技大学中山学院信息中心网络维护科的工单报修系统，依托于微信服务号提供服务。


技术栈：  
前端：Svelte/SvelteKit  
UI：[Carbon Componenets Svelte](https://svelte.carbondesignsystem.com/)  
HTTP通信：Axios  

后端：Go  
HTTP通信&杂项：Echo  
数据库：PostgreSQL（[sqlc](https://sqlc.dev)）

## 构建与部署
依赖：

```
npm

Go >= 1.24.6

GNU make 

```
在项目根目录下执行`make`，在`artifacts`里面查看后端可执行文件和前端素材文件夹，之后随便你怎么部署。

配置文件请查看`back/doc`下的示例文件，或者找开发组组长要一份生产环境下的。

这里附赠了systemd unit文件，想这样部署的话可以参考。

另外注意一下后端程序只监听`127.0.0.1`，是硬编码在程序里的，所以要套上一层反代。

后端可以自己托管前端文件夹，也可以在反代那里就把请求拦截下来，这个看怎么部署，根据你的需要，和喜好自行决定即可。
