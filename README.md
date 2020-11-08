# nets
Nets is a high performance HTTP restful API framework written in go (golang).

## 性能压测（wrk）

从三次压测结果来看，Nets的性能是要比Gin高一点的。

<b>对比框架</b>：Gin <br/>

<b>框架代码版本</b>：Nets版本为2020年11月3日最新代码，Gin为2020年11月3日最新发布版本 <br/>

<b>每次压测开始时负载</b>：1.0~1.05 <br/>

<b>硬件环境</b>： <br/>
![硬件环境](https://github.com/dlboY90/resources/blob/main/nets_wrk_mac_new.png?raw=true)

<b>Nets压测代码</b>： <br/>
![Nets压测代码](https://github.com/dlboY90/resources/blob/main/nets_wrk_nets_code.png?raw=true)

<b>Gin压测代码</b>： <br/>
![Gin压测代码](https://github.com/dlboY90/resources/blob/main/nets_wrk_gin_code.png?raw=true)

<b>Nets压测结果</b>： <br/>
![Nest压测结果](https://github.com/dlboY90/resources/blob/main/nets_wrk_nets_result.png?raw=true)

<b>Gin压测结果</b>： <br/>
![Gin压测结果](https://github.com/dlboY90/resources/blob/main/nets_wrk_gin_result.png?raw=true)
