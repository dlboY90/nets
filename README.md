# nets
Nets is a HTTP API framework written in Go (Golang). 

## 性能压测（wrk）

从三次压测结果来看，Nets的性能是要比Gin高一点的。

<b>对比框架</b>：Gin <br/>

<b>每次压测开始时负载</b>：1.0~1.05 <br/>

<b>硬件环境</b>： <br/>
![](https://github.com/dlboY90/resources/blob/main/nets_wrk_mac_new.png?raw=true)

<b>Nets压测代码</b>： <br/>
![](https://github.com/dlboY90/resources/blob/main/nets_wrk_nets_code.png?raw=true)

<b>Gin压测代码</b>： <br/>
![](https://github.com/dlboY90/resources/blob/main/nets_wrk_gin_code.png?raw=true)

<b>Nets压测结果</b> <br/>
![](https://github.com/dlboY90/resources/blob/main/nets_wrk_nets_result.png?raw=true)

<b>Gin压测结果</b> <br/>
![](https://github.com/dlboY90/resources/blob/main/nets_wrk_gin_result.png?raw=true)
