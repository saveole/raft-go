## Raft 算法学习实践
> 主要参考了Eli Bendersky 的 [Raft 系列](https://eli.thegreenplace.net/2020/implementing-raft-part-0-introduction/) 博客教程。

### Overview
- Leader Election
- Log Replication
- Safty

### 项目结构
- part1-election: 主节点选举实现


#### References
- [Raft](https://raft.github.io/)

#### 可视化

```go
go test -run TestElectionFollowerComesBack |& tee cb.log
go run ../tools/log2html/main.go < cb.log
```