### 集群选举机制
- 基本角色(对应状态机的状态)
    - 1.Leader
        ```
        1）接收客户端请求，将请求转发给Leader，Leader将请求写入日志，并返回响应；
        2）定时向集群发送心跳，保持Leader状态；
        3）当Leader挂掉后，Candidate将自身状态转换为Follower，并开始选举Leader。
        ```
    - 2.Follower
        ```
        1）接收 Leader 的日志，并更新本地状态；
        2）定时接受 Leader 发送来的心跳，保持Follower状态；
        3）当Follower挂掉后，Candidate将自身状态转换为Leader，并开始选举Leader。
        ```
    - 3.Candidate
        ```
        1）当Leader挂掉后，Candidate将自身状态转换为Follower，并开始选举Leader。
        ```
- 角色之间的转换过程


#### RoadMap

- [ ] using protobuf for serialization.
- [ ] visualize the log to demostrate the whole process.