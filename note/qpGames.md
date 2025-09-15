# 项目概述

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=NDUxNDQyZDYzM2FhM2IyMWY4N2QzNWQ5ZjUxZDQ0MGRfS214R2VqbnFxRjd0WU9QbVJLdFF6NGNwVURYdFo4QzVfVG9rZW46QTRmemJ0YmtBb2FlQ3R4bHNSNWNjdFpibm9nXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

> 高并发处理：
>
> - 协程池管理 ：每个服务最多支持10240个并发协程
> - 连接池优化 ：MongoDB连接池配置 minPoolSize: 10, maxPoolSize: 100 ，Redis连接池 poolSize: 10 ，有效复用数据库连接
> - WebSocket并发管理：能管理大量并发WebSocket连接
> - 消息队列缓冲：提供异步channel消息处理能力
>
> 横向扩展能力：
>
> - 微服务架构 ：实现基于Etcd的服务发现与动态负载均衡
> - NATS消息中间件 ：支持跨服务器消息路由，实现服务间解耦
> - 多实例部署 ：配置文件支持多个connector和game服务实例同时运行

## 项目介绍

qpGames是一个基于Go语言的分布式棋牌游戏平台，采用微服务架构设计，支持红中麻将和拼三张等多种游戏类型的实时对战。在技术架构上 ，该项目使用Gateway作为统一入口处理HTTP请求，Connector负责WebSocket长连接管理，Hall和Game服务处理具体业务逻辑，通过gRPC实现服务间同步通信，NATS消息队列处理异步消息； 数据存储采用MongoDB+Redis的组合，MongoDB负责持久化存储，Redis提供缓存和会话管理； 基于JWT Token机制实现用户身份验证和授权。

## **项目架构**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=OGIyMjdmODQ1MThlYmNkNzVhYWRjZTllZGM4NzJhNDNfRWphMzl0bTRaUEJwZEVuTllKOTIySFdKQk9rYjZyaFNfVG9rZW46TjdhU2JxV2lab2R1UjF4WmpjWGM4TmpGbmljXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

该项目是一个基于Go语言的分布式游戏服务架构 ，采用微服务设计模式，通过Gateway网关、Connector连接器、Hall大厅服务、Game游戏服务、User用户服务等核心模块，实现了多协议通信（HTTP/WebSocket/gRPC）和异步消息处理（NATS），数据存储采用MongoDB+Redis的组合方案，支持高并发实时游戏业务。

- 客户端 ↔ Gateway ：HTTP/HTTPS（RESTful API）
- 客户端 ↔ Connector ：WebSocket（实时通信）
- Gateway ↔ User ：gRPC（同步RPC调用）
- Connector ↔ Hall/Game ：NATS（异步消息队列）
- 服务发现 ：etcd（服务注册与发现）

## **请求流程**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=YmY1ZjQ5MjlhYWM2NGI2MzhkMzYwYjU4YzM3YmEwMjJfeDdnc1RPVTlCN1JnUjdwUVBMSkp4T1g3d1lQVllBZWxfVG9rZW46VHluS2I0Y3ZhbzN2OGl4R3ZXNGNHSVdlbkpHXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

- **用户注册**流程： 客户端通过HTTP POST请求向Gateway发起注册，Gateway通过gRPC调用User服务将用户数据持久化到MongoDB，然后返回JWT Token和Connector服务信息给客户端，完成用户身份认证和授权的建立。
- **游戏连接**流程： 客户端与Connector建立WebSocket长连接后，Connector通过JWT Token验证获取用户UID，随后查询Core服务层获取用户信息和游戏配置，最终为客户端建立有效的游戏会话Session。
- **游戏业务**流程： 客户端通过WebSocket向Connector发送游戏消息，Connector解析路由后通过NATS消息队列将请求转发给相应的Hall或Game服务进行业务处理和数据库操作，处理完成后通过NATS返回响应给Connector，最终通过WebSocket推送结果给客户端，实现完整的实时游戏交互闭环。

# 核心业务功能

## 用户系统

- **用户注册** 

网关接收用户注册请求后会先解析JSON参数，包括账号、密码、登录平台和短信验证码等信息。系统通过RPC调用用户服务的Register方法进行注册处理，成功后获取用户的唯一标识uid。接着构建JWT的Claims声明，包含用户uid和7天的过期时间，使用HMAC-SHA256算法和配置的密钥生成JWT令牌。最后将生成的token和连接器服务的地址端口信息一起返回给客户端，为后续的游戏连接做准备。

- **用户登录** 

客户端携带JWT令牌和用户基本信息发起连接请求。系统首先使用jwts.ParseToken方法验证令牌的有效性，解析出用户的uid标识。验证通过后调用用户服务的FindAndSaveUserByUid方法，如果用户不存在则自动创建新用户，设置初始金币数量、默认头像、昵称等信息并存储到MongoDB中。用户信息确认后将uid绑定到WebSocket会话中，同时返回用户完整信息和游戏前端配置数据，完成登录流程。

- **用户信息管理**

系统采用懒加载策略，首次访问时通过 UID 查询用户信息；如用户不存在则自动创建包含默认值的用户记录；支持更新用户基本信息（昵称、头像、性别）、地理位置和金币等数据；所有操作通过 MongoDB 进行持久化存储。

- **JWT****认证** 

JWT认证采用HMAC-SHA256签名算法，通过CustomClaims结构体定义载荷信息，包含用户uid和标准的注册声明。GenToken方法负责生成带签名的JWT字符串，使用配置文件中的密钥进行签名。ParseToken方法用于解析和验证JWT令牌，首先检查签名方法是否为HMAC类型，然后验证令牌的有效性并提取用户uid。系统在多个服务中都配置了相同的JWT密钥，确保令牌在不同服务间的一致性验证，过期时间统一设置为7天。

- **跨域处理**

实现自定义 Gin 中间件处理 CORS 跨域请求；配置允许的请求方法、头部和域名白名单；正确处理浏览器的 OPTIONS 预检请求，对于OPTIONS预检请求会直接返回204状态码，其他请求则继续后续处理流程。设置适当的缓存时间和凭据传递策略；确保跨域请求的安全性和兼容性。

## 游戏大厅

- **用户地址更新** 

前端通过大厅服务提交用户的地理位置信息，大厅接收到请求后解析JSON参数（包含address和location字段），然后调用用户服务层的UpdateUserAddressByUid方法更新MongoDB中的用户地址信息，最后返回包含更新数据的统一响应格式给前端。

- **游戏列表展示** 

系统支持多种游戏类型，包括拼三张、牛牛、跑得快、三公、红中麻将等，游戏类型定义在 `proto.go` 中。

## 房间管理

- **房间创建** 

网关会先检查用户是否合法，并把请求参数解析出来，确认这个用户是不是已经在某个房间里，如果在房间里就不允许再创建。接着找到用户所在的联盟，由联盟生成一个唯一的房间号并创建房间对象。房间在初始化时会根据选择的游戏类型加载对应的玩法逻辑，比如三张牌或者红中麻将。最后把房间登记到联盟的房间列表中，并让创建者自动进入这个新房间。

- **生成房间号**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=ZjkxZWMwMDU1YTI3YTllYmVmNGFjNjUwMWU2YmNmMDdfazFEc216RGo2NUZ6TUVOVHNTN0U0bEV4ZnMzZVNZVDZfVG9rZW46WWh0VmJjRXNmb2dkbE14bjh2S2NQQkJ4bmRiXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

系统首先调用 `CreateRoomId`，它会生成一个六位数的随机数作为候选房间号，先生成一个 `0~999998` 的随机数，如果小于 `100000` 就加上 `100000`，确保结果始终是六位数。拿到候选号后，`CreateRoomId` 会遍历检查该房间号是否已经存在，如果重复则递归重新生成，直到找到一个未被使用的号码为止。

- **加入房间**

用户通过房间号发起加入请求，系统先验证用户身份和房间是否存在，然后检查房间是否已满员或游戏是否已开始。验证通过后将用户添加到房间的用户列表中，更新房间状态，并向房间内所有玩家广播新用户加入的消息，同时返回当前房间的完整信息给新加入的用户。

*人数限制* ：房间最多支持6人参加，座位号从0-5共6个位置。不同游戏类型有不同的人数要求，通过GameRule配置中的MinPlayerCount（最小人数）和MaxPlayerCount（最大人数）字段控制。比如红中麻将需要达到MaxPlayerCount才能开始游戏，而拼三张等游戏达到MinPlayerCount即可开始。

*座位分配*机制 ：系统通过getEmptyChairID()方法自动分配座位号。分配逻辑是从座位号0开始遍历，找到第一个未被占用的座位号分配给新用户。如果房间为空则分配座位号0，否则按顺序查找空闲座位，确保每个用户都有唯一的座位标识。

- **房间解散** 

房间解散采用投票机制实现，当有玩家发起解散请求时，系统会记录该玩家的投票状态并向所有房间成员推送解散投票界面，显示各玩家的投票状态、头像、昵称等信息，并设置30秒倒计时。如果玩家选择同意解散，系统将其座位号记录到askDismiss映射中；如果所有玩家都同意解散，系统会依次踢出所有用户，清空房间用户列表，最后调用dismissRoom方法解散房间，该方法会取消所有定时任务并从联盟的房间列表中删除该房间。如果有玩家不同意解散，系统同样会推送当前投票状态，但不会执行解散操作。

- **自动踢人** 

通过定时器实现自动踢人功能，每当用户进入房间时会启动一个30秒的踢人定时器。如果用户在30秒内完成准备操作，定时器会被停止并删除；如果超时未准备，定时器会自动执行踢人逻辑。

踢人过程：首先向被踢用户推送空房间号的消息清除其房间状态，然后向房间内其他玩家推送通知有用户离开，最后从房间的users映射中删除该用户。如果踢人后房间变为空房间，系统会自动解散房间。整个机制通过goroutine异步执行，避免阻塞主要的游戏逻辑，同时支持重复进入时重置定时器。

## 游戏系统 

### 红中麻将

人数大于最大值可以开始游戏

- **牌型定义** 

HongZhong4模式 ：共 112张牌 （9×3×4 + 4 = 108 + 4 = 112张）**万筒条**

HongZhong8模式 ：共 116张牌 （9×3×4 + 8 = 108 + 8 = 116张）

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=Y2RhOTIxZDM2Zjg2MzhiZDg0MDk1MmE3YWMzZDJkMWRfeUhwVFlkeG9nSml6TVVIdVZGZ1Uyek9ITXdjZmpBd0FfVG9rZW46R3pTc2JTcm16b3FHZDZ4RGNHSGM5V0xubkhmXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

#### **洗牌**

构建包含所有牌的数组，然后使用多轮乱序算法进行洗牌，通过300次循环随机交换不同位置的牌来打乱顺序。

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=ZjA0MWM0NDU3M2ZjMzhiMTc1YTBhNjE3ZDI3OGNjNTFfM3hjYUlsQW4zaUgwNG82dnl4bU5GV01EdkxIMmlIYXhfVG9rZW46SVlvUWI1emNnb3p5WXN4NThwVGN1eWFlbm9lXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

*当前算法缺陷*：

1.非等概率分布 ：由于使用 i % len(l.cards) 作为索引，某些位置会被访问更多次（300次循环中，前面的位置会被重复访问）

2.偏向性 ：牌的最终分布会偏向某些特定的排列组合

3.效率低下 ：300次交换操作过多，造成不必要的计算开销

4.理论缺陷 ：多次随机交换并不能保证每种排列的概率相等

*优化方案：*

Fisher-Yates 洗牌算法

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=NjU4NzJlMGU1MGEzYWM5YmUxYWY5NDRlZTEyM2IzOGNfWk9ic014cHhaR2djSXgybk8xY0U0QkU2Mmc1UHJ6c2hfVG9rZW46SW1wcWJsZ2w2bzRDQWZ4OWNwSWNBZ3JWbmdjXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

> 原理：
>
> 从最后一张牌开始，与前面（包括自己）的随机一张牌交换
>
> 每次交换后，该位置的牌就确定了，不再参与后续交换
>
> 保证每张牌出现在任意位置的概率都是 1/n

> 优势：
>
> 1.严格等概率 ：每种排列出现的概率完全相等（1/n!）
>
> 2.时间复杂度 O(n) ：只需要 n-1 次交换，比原来的300次更高效
>
> 3.有严格的数学证明保证等概率性
>
> 4.被广泛采用的标准洗牌算法

- **发牌**

为每位玩家发13张牌（从洗好的牌堆里拿出前 13 张），发牌时会为每个玩家构造一个只包含自己手牌、其他人手牌全部隐藏（用36表示）的数据结构，然后推送给对应玩家。

- **游戏操作** 

游戏操作通过OperateType枚举定义，包括吃胡（HuChi）、自摸（HuZi）、碰（Peng）、吃杠（GangChi）、补杠（GangBu）、自摸杠（GangZi）、过（Guo）、弃（Qi）、拿牌（Get）等操作类型。根据不同操作类型执行相应逻辑，如碰牌时删除手牌中对应的2张牌并记录操作，杠牌时删除3张牌，胡牌时将牌加入手牌并调用gameEnd结束游戏。每次操作都会通过GameTurnOperatePushData向所有玩家推送操作信息。

#### **胡牌**

胡牌判定方法：*预计算查表+**递归**鬼牌尝试*

胡牌=4*（顺子ABC/刻字AAA）+1*将（BB）

系统启动时预生成所有胡牌组合，编码为9位数字字符串存储在分层字典中。实际判定时将手牌按花色统计编码，通过递归函数（3+3+3+2+3 5个层级 ，最多递归5次就可以） checkCards 逐花色查表，当鬼牌不足时自动增加鬼牌数量重试，直到找到有效组合。胡牌必须满足（牌数+鬼牌数）%3==2的条件，确保有一个对子作为将。

编码：[1,1,1,2,3,4,5,6,7,8,9,9,9]  ->"311111113"

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=NmIyNmEwYWJmZTE0NWI4ZTM3OTM0YzVlZjYwZDI0MWVfREE1b1JLNE5CdE50WlBVQm14S3pCc0JzZFJraFl2RzhfVG9rZW46QTBFaGJaakxKb3Z0WFR4Ujd6YmNlVTIybm9lXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

> **缺陷**

- 预计算所有可能组合需要大量内存存储字典表
- 程序启动时需要生成完整的字典表，耗时较长
- 递归深度过深 ： `genGui` 方法递归生成鬼牌组合，可能导致栈溢出
- 算法与特定麻将规则强耦合，难以支持其他变种

> **优化**

- 算法层面：
  - 动态规划替代查表法 - 用按需计算替代预计算全表，大幅减少内存占用和启动时间
  - 位运算加速计算 - 用位操作替代字符串编码，显著提升比较和计算效率
  - 迭代替代深度递归 - 避免栈溢出风险，提升算法稳定性
- 性能：
  - 懒加载字典表 - 按需生成缓存，大幅缩短程序启动时间
  - 并发处理花色检查 - 多核并行处理，充分利用CPU性能
  - 对象池复用内存 - 减少GC压力，提升性能稳定性

- **结算**

当玩家执行胡牌操作时自动触发结算机制。系统首先验证操作记录的有效性，确保最后一次操作确实是胡牌行为，然后构建 `GameResult` 结构体包含详细的结算信息，包括各玩家得分数组、所有玩家手牌、剩余牌堆、胜利玩家ID数组、胡牌类型、码牌信息和放杠数组等数据，最后通过GameResultPushData将结算结果推送给房间内所有玩家，并设置3秒延时后自动调用resetGame重置游戏状态准备下一局。

- **吃胡**

别人打出一张牌，你刚好能用这张牌胡牌

当有玩家出牌后，系统检测每个玩家是否可以对刚打出的牌进行操作。首先验证牌的有效性（card值在0-35范围内），然后调用 `canHu` 方法判断玩家当前手牌加上这张牌是否能够胡牌，如果可以胡牌则将HuChi操作添加到可操作数组中，同时自动添加Guo（过）选项让玩家可以选择放弃胡牌。

- **补杠**

已经执行过碰牌操作，当前摸到相同的牌，则可以进行补杠

补杠处理分为两种情况：*自摸补杠*发生在玩家自己的回合中，系统会向所有玩家推送操作信息但只向执行补杠的玩家显示真实牌面，其他玩家看到的是暗牌，补杠后从手牌中移除4张相同的牌并记录操作，然后通过 `setTurn` 函数让玩家继续摸牌；*吃牌补杠*是特殊情况，在某些麻将规则中不被允许，代码中已被注释掉以避免争议。

- **超时自动出牌** 

到时出牌通过定时器机制确保游戏连续性。玩家轮到出牌时启动30秒倒计时，系统每秒检查一次倒计时状态，将剩余时间推送给客户端显示。倒计时归零时自动执行操作：优先选择弃牌，其次选择过牌，确保游戏不会因玩家未操作而卡住。每个座位维护独立定时器，避免多玩家操作冲突。

### 拼三张

人数大于最小值可以开始游戏

1. #### **牌型定义**

> 十六进制编码:
>
> 方块 ：0x01-0x0d（十进制1-13）
>
> 梅花 ：0x11-0x1d（十进制17-29）
>
> 红桃 ：0x21-0x2d（十进制33-45）
>
> 黑桃 ：0x31-0x3d（十进制49-61）
>
> - 共52张牌

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=Y2VmNDAxYzcyYzBjM2RiMmQ2MGE0Yjc3MTU4MjUyZmJfcWJ5SFpRdXVzcjNGZ2RxaTRBZDFxVzhlaldVNWt6dUJfVG9rZW46QTVSMGJtS2t3bzl6UVd4MjBBQ2NCdkpUblFlXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

- **获取牌花色和点数**

*提取点数*：

card & 0x0f  // 位与运算获取低四位

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=ODAzZDZlN2IwZWMzNjkwZjdiMDg4NTc5NjIxMGQwYWNfZ3BnNkl1aWo5azdweXZrZW5saXNVYk9kMkd6ODFQVTFfVG9rZW46WktsMWJmTFZqb1VSVDF4bWNpV2NaYktpbjNkXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

*提取花色*：

card / 0x10  // 整数除法获取高四位

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=YWI3YWU3M2FmYTRlYmYyMjkzMmQwZTBmZWJmZWVkNzhfZFlSYXNvcGdOSUFsWWgzZFlHS09FeWRNcVp6MnpRWlhfVG9rZW46TFl5N2JLeDVKb0R0TmZ4V042UmNRN1dtbkdiXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

1. #### **洗牌**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=MjAzYWQ4ZThiNjE3NzQ0ZjYwYjc4NGVjZDdmZjZiYzFfVWpwQzZhVzFYRkUwRkF1Vk1IYkcxR01lcHdCR1kyUENfVG9rZW46VTVQN2JKTjAwb2o2d094c2ZRQ2NSMjlvblBjXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

> 缺点：**随机性不完全均匀**。 因为每次都是在 `0 ~ len(cards)-1` 范围里随机，而不是“逐渐缩小范围”。

*优化：Fisher–Yates 洗牌*

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=MWM2MDZlNGM5MTFiZTJkY2FhZmVmOWI4NDU3YTBjMzRfUUxGSDZLdnYyRmFWbzdZUXg5TGxabk13b1FHaU9SVkxfVG9rZW46VWE1MmJhZlNLb0JCclh4T1VWNGNaVnhWbjFuXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

> 原理：
>
> 从最后一张牌开始，随机选择 `[0, i]` 的一张牌和它交换。
>
> - 下一步就只考虑前 `i-1` 张牌，直到处理到第一张为止。
> - 这样保证每张牌在最终任何位置的概率相等。

1. #### **比牌**

*牌型大小比较*：

豹子 > 顺金 > 金花 > 顺子 > 对子 > 单牌

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=Y2IxZjU3YWU5ODJkMzNjZWUxMzAyNjFjOGY5ZmM4ZGJfV0RwVnJEZ3l3d1gyQ1RLdmxrYkRjSW1oamlnUjJMZXFfVG9rZW46UVoxbmJtS1Rkb2lBbkF4YkZtZWNvc3JVbkJnXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

1. 豹子：三张相同点数牌，最大
2. 顺金：既是顺子又是金花
3. 金花：同花色
4. 顺子：连续点数，支持A23特殊顺子（特殊处理了A牌，将点数1变为14）
5. 对子：两张相同点数
6. 单牌：散牌，最小

> 比牌流程：
>
> 先比牌型大小 → 豹子 > 顺金 > 金花 > 顺子 > 对子 > 单张。
>
> - 牌型相同 →
>   - 如果是对子，先比对子点数，再比单牌。
>   - 如果是其他牌型，逐张比点数（最大 → 次大 → 最小）。
> - 全部相同 → 平局（平局时主动比牌方判负）

比较两个玩家的手牌大小，处理比牌结果（平局时主动比牌方判负），更新胜负玩家的状态为 Win 或 Lose ，将败者加入 Loser 列表、胜者加入 Winner 列表，推送比牌结果后继续游戏流程。

- **看牌**

首先验证游戏状态必须为下注阶段且轮到当前玩家操作，然后将玩家的看牌状态记录到 LookCards 数组中，更新玩家状态为 Look ，并分别推送给看牌玩家（包含具体牌面）和其他玩家（仅通知看牌事件）不同的消息内容。

- **下注**

验证游戏状态、当前操作玩家和下注金额的合法性后，将下注分数记录到 PourScores 二维数组中，计算当前玩家总下注和全场总分数，推送下注信息给所有玩家。

- **弃牌**

验证玩家状态后将其加入失败者列表，将其他仍在游戏中的玩家加入胜利者列表（要继续进行下一轮比牌），设置玩家状态为 Abandon ，推送弃牌消息给所有玩家，后调用结算逻辑检查游戏是否结束。

# 技术栈

## WebSocket

- **连接管理**

`Manager`连接管理器（维护clients映射表）负责WebSocket连接的建立、维护和销毁，统一管理所有客户端连接：

> - 建立：将HTTP请求升级为WebSocket，并创建对应的连接对象。
> - 维护：将连接纳入统一管理，配合*读**写锁*确保在高并发场景下的线程安全。
> - 销毁：当连接断开或异常时，安全地移除并释放资源。

**HTTP 升级为 WebSocket 连接**：通过 HTTP 协议的 *Upgrade* 机制完成

在客户端（通常是浏览器）第一次发起请求时，会先走普通的 HTTP 请求，在请求头里带上：

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=ODExYjU1M2FkMzMzYmI2MDEwNjcyODNiNTU4ZDgxNGNfZVFsOXozdlhNQWFMZkV4Q3RFRWRMNWlsM24wNFVzV1NfVG9rZW46RE9oTGJDSkhNb1k1V1h4clZXN2NRNzhQbmdmXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

服务端收到后，如果支持 WebSocket，就返回一个 *101* Switching Protocols 的响应，表示协议切换成功，同时返回经过加密计算的 `Sec-WebSocket-Accept` 字段，确认握手有效。

握手完成后，HTTP 协议就被升级为 WebSocket 协议，接下来双方就不再走 HTTP 请求-响应模式，而是通过持久的 *TCP* *通道进行**全双工通信*。

在 Go 里用 `websocket.Upgrader.Upgrade(w, r, nil)`，完成从 HTTP 到 WebSocket 的协议切换。

- **心跳检测与连接保活**

采用Ping/Pong机制，服务器定时发送Ping帧，客户端收到后自动回复Pong帧，通过双向确认检测连接的活跃状态

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=MzQ1Y2Y5M2UyZTExZjE2NWEwMjEzMmFlNWNjYmUwN2FfRmsyd0N4dTBCM09WWUdEMmtyZmhnbEtua3I4SU1HWmlfVG9rZW46QVdzcmI4S0ltb2U4ekx4ZW9zV2NJbjFYbnljXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

> **pingInterval < pongWait**，保证在连接被判定超时前，一定能发出至少一次 Ping，给客户端充足的响应机会。
>
> 预留 **10% 缓冲时间**，避免因网络抖动或延迟造成误判。

> - **服务器端主动****Ping** ：通过定时器每9秒主动发送WebSocket Ping消息，设置10秒写入超时，一旦发送失败立即关闭连接，确保服务器能够及时发现断开的连接并进行清理。
> - **客户端Pong响应** ：每当收到客户端的Pong响应时自动重置读取超时时间为10秒，收到响应就延期,确保活跃连接不会被误判为超时，实现了连接状态的动态维护。
> - **应用层心跳消息** ：HeartbeatHandler在WebSocket协议层Ping/Pong机制之上提供应用层心跳服务：接收客户端心跳数据包 -> 构造空的响应数据并编码 -> 将心跳响应发送回客户端
> - **连接超时检测与自动清理** ：当连接在10秒内未收到任何消息（包括Pong响应）时自动触发超时断开，配合主动Ping机制形成完整的双向检测体系，确保异常连接能够被及时发现和清理。

- **用户会话管理**

用户会话管理功能为每个WebSocket连接维护用户会话信息，存储用户状态和临时数据，支持本地和远程会话管理。

*本地会话*：为每个WebSocket连接维护独立的会话实例，包含连接ID、用户ID和会话数据存储，提供了创建会话、单个数据操作、批量数据设置等核心方法，所有操作都通过*读**写锁*保证线程安全。

*远程会话*：实现跨服务的会话数据同步，支持向指定用户推送消息和实时会话数据同步，内部使用NATS消息队列进行服务间通信。当远程服务的会话数据发生变更时，会通过NATS将数据推送到目标服务，目标服务更新本地连接的会话数据，确保分布式环境下的数据一致性。

> **稳态会话**：保持连接持续活跃且数据一致的会话状态
>
> 实现机制 ：
>
> - 心跳检测 ：的Ping/Pong机制
> - 会话同步 ：实时推送会话数据
> - 连接恢复 ：通过NATS消息队列保证跨服务会话数据同步

## Pomelo

Pomelo协议是项目*客户端与服务端*实时通信的核心协议，基于WebSocket传输，专为游戏场景优化。

- 支持握手(Handshake)、心跳(Heartbeat)、数据(Data)、踢出(Kick)四种包类型，以及Request/Response、Notify、Push四种消息类型，满足游戏中不同的交互需求。
- 采用路由压缩机制将字符串路由映射为数字编码，使用变长编码减少网络传输开销，支持GZIP数据压缩，显著提升通信效率。
- 实现完整的连接生命周期管理，包括握手协商、心跳保活、异常处理和优雅断开，确保连接的稳定性和可靠性。

> **原理**：
>
> - Pomelo协议基于WebSocket传输层，采用二进制消息格式，通过标志位区分不同的包类型和消息类型，实现客户端与服务端的双向实时通信。
> - 使用变长编码(Varint)压缩数字数据，支持路由字典将字符串路由映射为数字编码，减少网络传输开销。消息结构包含flag、messageId、route、data等字段，支持GZIP压缩。
> - 实现握手协商机制建立连接参数，通过心跳包保持连接活跃，支持服务端主动推送和踢出客户端，提供完整的连接生命周期管理。

> **好处：**
>
> - 性能优化 ：路由压缩和变长编码显著减少网络带宽消耗，二进制格式提升解析效率，适合高频实时通信场景。
> - 灵活性强 ：支持等多种消息模式，满足游戏中不同的交互需求，如玩家操作、状态同步、实时推送等。
> - 可靠性高 ：内置心跳机制检测连接状态，支持错误处理和异常恢复，握手协商确保客户端服务端参数一致性。
> - 易于扩展 ：模块化设计便于添加新的消息类型和处理逻辑，路由字典机制支持动态路由管理，适应业务发展需求。

## NATS 

通过*发布订阅*模式实现connector、hall、game等各服务间的异步通信、消息路由和负载均衡，为整个微服务架构提供低时延、高可用的解耦通信基础设施。

- **跨服路由：基于****服务类型****的智能路由和****负载均衡**

> **路由解析**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=MmRjYTIxZmFkMzQ0MDY3NzVlYmVlNjAzMmVkMGYwNzdfNGxLMGQ5SElwQ0xBaDVnS1VSeVE0M0ZSa1FuVzh4a0NfVG9rZW46TGVnSmJVMlc2b2VCRnd4UlNyOWN4d3FIbk1iXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

路由格式： serverType.handler.method（服务类型.处理器.方法 -> eg:hall.entryHandler.register）

> 检查是否为本地处理 
>
> - 本地处理：执行业务逻辑
> - 远程处理：通过 NATS 转发到目标服务器

> **随机****负载均衡**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=YWQ5NTkyNDY4NzdiNmZjZDVjZmFhYTZjZTY5MDg1MzBfcG9lU2k3UVBzZ0VRZ1cyMVlMRHptYjNzdklkVHVFOU5fVG9rZW46SWxJZWIwdGJhbzFQTUx4YUV2SGNYd0FFbkFnXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

在指定类型的服务器列表里随机选出一个服务器，并返回它的 ID，如果没有该类型的服务器则报错

- **低时延推送：异步****Channel** **+** **发布订阅****模式的****毫秒****级消息传输**

> 异步Channel

通过channel机制实现非阻塞读取nats消息。当消息到达时，NATS客户端立即将消息投递到缓冲通道（容量：1024）中，避免等待处理完成

> 发布订阅模式

每个服务以自己的服务ID（connector001、hall-001、game-001）作为主题订阅消息，消息发布者直接向指定主题发布消息，无需等待确认

- **服务****解耦****：各服务独立部署通过NATS****松耦合****通信**

通过配置文件独立管理各服务实例（connector、hall、game等），每个服务只需知道目标服务的ID而无需了解具体的网络地址和端口。服务间通信完全通过NATS的主题路由机制，发送方将消息发布到目标服务的主题，接收方从自己的主题订阅消息。同时通过统一的Client接口抽象，屏蔽了底层通信细节，服务只需调用SendMsg方法即可与任意其他服务通信，实现了松耦合架构。

### **NATS** **Kafka** **RabbitMQ对比**

1. **性能和延迟**

- NATS：专注于低延迟和高吞吐量，微秒级延迟，适合高实时性需求场景。在小消息场景中表现非常出色。
- Kafka：通常用于大数据和日志处理场景。延迟较高，但能实现高吞吐，适合大规模流数据处理。
- RabbitMQ：在消息确认机制的支持下可以确保消息传递的可靠性，延迟和吞吐量较低，但非常适合处理可靠性和灵活性要求高的消息流。
- **消息模式**

- NATS：支持发布/订阅和请求/响应模式，适合实时通信和微服务通信。
- Kafka：以发布/订阅模式为主，专为流处理和事件溯源设计，且支持消费偏移管理。
- RabbitMQ：支持发布/订阅、点对点、主题等多种路由模式，适用于消息队列的灵活路由需求。
- **持久化**

- NATS：不支持持久化，NATS JetStream扩展了持久化功能，但仍侧重于短期消息存储。
- Kafka：强大的持久化能力，所有消息都存储在磁盘上，适合长时间数据存储和流式处理需求。
- RabbitMQ：可选择性持久化消息，持久化会影响性能，适合需要可靠消息传递但不要求大规模持久化场景。
- **可靠性和消息确认**

- NATS：轻量级且无状态，提供消息确认机制，可靠性不如RabbitMQ。
- Kafka：高可靠性设计，通过复制和日志存储保证消息不丢失，适合对消息持久性要求高的场景。
- RabbitMQ：支持消息确认、重试和持久化机制，可靠性最高，适合对消息交付顺序和持久性要求高的业务。
- **应用场景**

- NATS：适用于实时性要求高的分布式系统、微服务和事件驱动架构，特别适合小数据量的高频消息传输。
- Kafka：广泛应用于大数据实时分析、事件流处理、数据管道、日志聚合和事件溯源等。
- RabbitMQ：适合复杂的工作流、企业系统间通信、需要高可靠性和消息顺序的应用，如订单处理。

> **总结**

- **NATS**：专注低延迟、实时性强，适合微服务和高频实时通信。
- **Kafka**：高吞吐、强持久化，适合大数据和流处理。
- **RabbitMQ**：可靠性高、支持复杂路由，适合企业应用和事务处理。

## Etcd 

- **服务注册中心**

注册流程：

1. 创建注册器实例（传入etcd客户端、服务信息（名称、地址、权重等），设置租约TTL（默认10秒））
2. 构建注册Key

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=MzAyOTMyODVkNzQzZTBkYjAzNjM2NjAxYTZiYTUyNmFfemZzVFJwOTd3Y0NWSUoyWG9zU25Takh3aEU3TmxoakhfVG9rZW46Sm9XQ2JkTjM4b0tKYmF4cFduVGNBdEZ1bnBoXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

1. 创建租约
2. 绑定服务到租约（将服务信息JSON序列化作为Value -> 使用 Put 操作将Key-Value绑定到租约 -> 租约到期时，服务记录自动删除）
3. 启动心跳续租
4. 心跳监控（每隔TTL/3时间发送心跳（约3.3秒）->监听心跳响应，确保租约有效 ->心跳失败时记录错误日志）

服务注销流程：停止心跳 -> 撤销租约 -> 清理资源

## gRPC 

- **服务发现**

采用 etcd作为注册中心 + gRPC自定义Resolver 的架构实现：

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=M2E3NTViZThmZWY2NmMzZmMzZjU5ZGZhZWI5MDExNDNfWXR2bGZSME9KbDM0aVcwRE1yMWl5OWpyNnBuNGNaSmFfVG9rZW46THVRYWI3UUNXb1NyNER4Z2V1NWNQMThZbmdiXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

服务启动时通过Register组件将自身信息（地址、权重等）以租约形式注册到etcd，并启动心跳续租保持在线状态；

*gRPC**服务发现*流程：客户端解析etcd://地址后调用自定义Resolver.Build方法建立etcd连接并同步服务列表，然后启动Watch监听实时捕获服务变化事件并自动更新负载均衡池，实现服务发现和故障转移。

- **负载均衡**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=M2FkZjg4MmY4Nzk2ZThkYzIwZTJkZjZkNzY2ZGMyMjhfZkxGZGYyMzZQTno1c3RGZ3MyVW95OUhhdHUxNjl1enFfVG9rZW46WFdjWWJvTFRab1hZN3N4TXo5TWNwR2g0bnZoXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

> - 采用gRPC内置的 round_robin 轮询算法
> - 工作原理 ：按顺序依次将请求分发给不同的服务实例
> - 确保每个服务实例获得**相等**的请求分配

## Redis

- **用户ID生成器**

利用Redis的INCR命令，以"MSQP:AccountId"为键从10000开始递增生成全局唯一的用户ID（只存储最大值）

使用了Redis的*String*数据类型来存储数据

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=ZWU3ZmZlZjgwNmVhMTZiNjVmMzgwY2FhMGFiOTUyODFfc01aQm1PeUpYQW1MRTBrcUowalQ3RWlld3NoQ0J4QUNfVG9rZW46UmRYWmJ6ZTU2b0tzSEN4V2luNGNuNWtFbm1mXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

- **数据缓存**

> 当前项目 ：单机Redis + 分布式应用架构（多个应用服务connector、game、hall等共享同一个Redis缓存）

通过RedisManager封装单机和集群两种Redis客户端，提供统一Set接口实现键值对存储和TTL过期管理

> 键："MSQP:AccountId"
>
> 值："10086"（当前最大用户ID）
>
> 过期时间：0（永不过期）

## MongoDB

MongoDB在项目中作为核心数据存储引擎，通过BSON映射机制实现用户信息、账号认证、游戏数据等业务数据的*持久化存储*和高效访问。

> BSON映射：
>
> - 序列化：Go结构体 → BSON文档 -> InsertOne(ctx, user)
> - 反序列化：BSON文档 → Go结构体 -> singleResult.Decode(user)

user集合：存储用户的完整游戏档案，包括个人信息、游戏资产、社交关系和联盟数据等所有业务相关数据。

> - `FindUserByUid` ：根据UID查询用户
> - `Insert` ：插入新用户
> - `UpdateUserAddressByUid` ：根据UID更新用户地址信息

account集合：存储登录认证信息，通过Uid与User集合建立关联。

> - `SaveAccount` ：保存账号信息

> **为什么用mongo不用****mysql**
>
> - 游戏用户数据包含邀请消息数组、联盟信息等复杂结构，MongoDB文档型存储天然支持
> - MongoDB无需预定义表结构，易于扩展新字段
> - 游戏实时数据更新频繁，MongoDB写入性能更优
> - 一次查询获取完整用户信息，避免MySQL多表JOIN
> - 无需设计复杂表关系，开发速度快
> - mongo和redis都属于NoSQL生态，与Redis技术栈一致性更好

## Gin 

- **HTTP** **API网关**

Gin作为整个微服务系统的HTTP入口 ，通过 gin.Default() 创建Gin引擎实例，使用 r.POST() 注册路由端点，并通过 r.Run() 启动HTTP服务监听客户端请求。 Gin提供了完整的HTTP服务器框架 ，让系统能够接收和处理来自前端的HTTP请求，实现了Web API的基础设施。

- **JSON****数据处理**

Gin提供了数据绑定和序列化能力 ，通过 ctx.ShouldBindJSON() 自动将HTTP请求中的JSON数据解析并*绑定*到Go结构体，通过 ctx.JSON() 将Go对象*序列化*为JSON格式返回给客户端。 

- **协议转换桥梁**

Gin充当HTTP协议与gRPC协议之间的转换器 ，接收HTTP请求后调用后端gRPC服务，并将gRPC响应转换为HTTP响应返回给客户端。 (Http->gRPC   gPRC->http)

## JWT

> JWT（A.B.C） A:定义加密算法 B:存储数据 C:签名

JWT采用*HMAC-SHA256*签名算法和*7天*有效期设计

用户完成注册后，Gate服务立即生成包含用户唯一标识（UID）的JWT Token作为身份凭证

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=NGQzMzc2NjNjMWIyMjQ3N2Y3MGJjYzczOTYxYjRhNjhfNHZ6c2VYdVFxTUdaUHJLMXFzSkZmTUhzeXprNnI1Q1VfVG9rZW46SWM5bGJibHNsb1pRb0J4dVo3RGNTMTFObkZCXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

当用户尝试连接游戏时，Connector服务通过验证JWT Token确认用户身份并建立WebSocket会话

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=ZGZmNWU3NTU1N2UzYjhlNTZkMTgzNDcyZTg4NDAzMGJfMG9YM1E0WTBmZWdtSXJrMEhyaHp0U0ZXUko0MWdGczBfVG9rZW46VUc0V2JTMVdUb2pianR4QlVpaWNORDRTbkpnXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

## CORS

解决前端Web客户端与后端API服务以及WebSocket服务的跨域访问问题，确保不同域名下的游戏客户端能够正常与服务器进行HTTP请求和实时通信。

- **前端Web客户端 ↔ 后端****API****服务跨域处理**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=NWQ1NTJiYWViZmRiYjJmZDdhZDRmYmRiNWMyZDdhOWRfYjd5WWVsYlU2cWdiVlVlOWpYMWtXVk9hSXN1eTFkWjlfVG9rZW46RFBRN2JNMU1hbzR1SzN4MWtSM2NWWWhTbnplXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

检查请求的Origin头部，动态设置CORS响应头来允许跨域访问：设置 Access-Control-Allow-Origin 为请求来源域名，允许POST、GET、PUT、DELETE、OPTIONS等HTTP方法，允许Content-Type、Content-Length、Token等请求头，暴露Token响应头给前端，设置预检请求缓存时间为48小时，允许携带认证凭据，并特别处理OPTIONS预检请求返回204状态码，从而确保不同域名的前端应用能够正常调用后端API接口。

- **前端Web客户端 ↔ WebSocket服务跨域处理**

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=ZDg4ZWZkMzc0Y2MxODAzYmU0YzViZDJiNmJmNzk5NTdfVlNFS2pDUHNPcDdMOFVNY1FpdHZ5OU1tYVdNNjZBQnhfVG9rZW46Tm5VbWJzOERNb1R6QXN4d2tWU2NSTXJrbnJoXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

通过CheckOrigin函数直接返回true，允许所有域名的WebSocket连接

## Protobuf

- 通过 `user.proto` 定义消息结构，*自动生成*gRPC客户端和服务端代码 （`user.pb.go` 和 `user_grpc.pb.go`*）*，为gate网关和user服务提供*类型安全的**RPC**调用接口*，实现高性能的微服务通信。
- 采用混合*序列化*策略：gRPC服务间通信使用protobuf原生序列化 （高效、类型安全）， WebSocket和游戏消息使用JSON序列化 （灵活、易调试）。

> **Protobuf**

Protocol Buffers是一种语言无关、平台无关的可扩展*序列化**结构数据*的方法。它类似于XML和JSON，但更小、更快、更简单。

- **体积小** ：序列化后的数据比JSON小3-10倍，比XML小20-100倍
- **速度快** ：序列化和反序列化速度比JSON快20-100倍
- **内存****占用少** ：运行时内存使用更少
- **支持多种编程语言**，一次定义，多语言使用，保证不同语言间数据交换的一致性
- **向后兼容**：新版本可以读取旧版本数据，字段可以安全地添加、删除或修改，保证系统升级时的数据兼容性

## Viper+fsnotify 

Viper配合fsnotify用于*配置文件的动态加载和热更新* ，实现了配置文件修改后无需重启服务即可生效的功能

> - 使用Viper读取YAML和JSON格式的配置文件
> - 通过fsnotify监听配置文件变化，自动重新加载配置
>
> YAML格式 ：用于应用配置（application.yml），包含数据库、服务端口、日志等配置
>
> JSON格式 ：用于游戏配置（servers.json），包含服务器集群、连接器等配置

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=OTdhYmIxMTA0NTBmYTdkZjFkYjAxZjdiZGM0MzRlYWZfTVFlQ2NCOGZxb1ByelNycDhLeXJybjNLTndrcHIxd1FfVG9rZW46QThyV2JWUHo2b09Pc214ZkVBQ2NLMWZpbkhFXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

Viper工作流程：

创建Viper实例，设置配置文件路径和格式，根据文件扩展名自动选择解析器读取配置内容，通过反射将配置数据映射到Go结构体，同时启用 WatchConfig() 集成fsnotify监听文件变化，当配置文件修改时触发 OnConfigChange() 回调函数重新解析配置并原子性更新内存数据。

fsnotify在底层使用操作系统的文件系统事件（如inotify、kqueue等）监听文件变化，当配置文件被修改、保存时，fsnotify捕获到文件变化事件，触发OnConfigChange回调函数，重新读取和解析配置文件，将新的配置数据更新到全局配置变量中，应用程序立即使用新的配置，无需重启服务

## statsviz 

statsviz 是一个 Go 应用程序的*实时可视化监控*工具，提供 Go 运行时的内存使用情况、垃圾回收统计等关键指标的可视化展示     

通过 HTTP 服务器提供 Web 界面访问，默认访问地址为 http://localhost:5854/debug/statsviz/                          

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=MDcwMTc5ZDRkOWM2ZTAwOTMwNTkwMGM1Mjc1ZWU4MzhfakQ0bmZFMld1OTN5WG5Ld1NQZnNxRWttZ21zSEpkWEhfVG9rZW46WVNuNGJtVnpqb3NzQVF4UFV2aGN6cDFWbmtlXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

在每个服务的 main 函数中通过 goroutine 异步启动监控服务

![img](https://qx9auby2jc5.feishu.cn/space/api/box/stream/download/asynccode/?code=ODA3Nzg0ZDZhNjYwZWQ4YTVjZGI4MDQ3ZjVlZWRhNzBfNXRxdGQ0dGN1TlJMZGdYVjVTb2s2Tld2dVpnOXJsZVVfVG9rZW46TFhFMWJQMmhWb3owNnV4WEYwTWN1aDlmbmZjXzE3NTc5MjI1Njc6MTc1NzkyNjE2N19WNA)

> 作用：
>
> - 实时观察应用程序的内存使用、GC 频率等性能指标
> - 快速定位内存泄漏、性能瓶颈等问题
> - 提供直观的应用程序健康状态监控界面

# 面试问题

1.**为什么要写这个项目**

为了深入学习分布式系统设计，掌握Go语言在高并发场景下的编程技巧和性能优化，实践微服务架构的拆分、通信和治理，积累游戏后端开发的业务经验和技术栈，探索实时通信、状态同步、负载均衡等核心技术在实际项目中的应用，提升对复杂系统架构设计和问题解决的综合能力。

2.**讲讲****分布式**

- 架构：将单体应用拆分为多个独立的微服务；每个服务负责特定的业务功能（用户管理、游戏逻辑、网关等）；服务间通过gRPC进行通信；使用etcd实现服务注册与发现
- 优势：
  - 服务独立部署，提高系统可用性
  - 技术栈多样化，每个服务可选择最适合的技术
  - 水平扩展能力强，可根据负载独立扩展服务

3.**项目亮点**

- 智能胡牌算法优化 ：采用预计算查表法将胡牌判断复杂度优化到O(1)，通过穷举法生成完整胡牌编码表，大幅提升游戏响应速度和用户体验。
- 高性能实时通信架构 ：基于WebSocket+NATS消息队列构建的实时游戏通信系统，支持毫秒级的状态同步和消息推送，通过连接池和消息路由优化实现高并发处理。

4.**项目难度最大的地方，是怎么解决的**

- 胡牌算法的性能优化 ：实时判断胡牌需要复杂的组合计算，直接计算会影响游戏体验

解决方案 ：

- 预计算策略 ：启动时生成所有可能胡牌组合的查找表
- 编码优化 ：将牌型编码为紧凑的数字表示
- 内存换时间 ：将查找表加载到内存，O(1)时间复杂度查询
- 支持鬼牌 ：预计算0-7个鬼牌的所有胡牌可能性

采用"N 连子+M 刻子+1*将"的标准麻将胡牌规则，使用穷举法预生成所有可能的胡牌组合并存储为查找表，将牌编码为字符串作为查表键值，通过查表法实现O(1)时间复杂度的胡牌判断。

5.**项目架构，怎么发牌，怎么码牌**

- 采用微服务架构：
  - Gateway层 ：使用Gin框架提供HTTP API，负责用户注册登录和JWT鉴权
  - Connector层 ：WebSocket网关服务，处理客户端长连接和消息路由
  - Hall层 ：大厅服务，处理房间管理和用户匹配
  - Game层 ：游戏逻辑服务，处理具体的麻将、三公等游戏逻辑
  - User层 ：用户服务，使用gRPC提供用户相关功能
- 发牌：采用多轮乱序算法确保随机性，循环 300 次，每次取当前轮数对应的位置（`i % len(l.cards)`）和一张随机位置的牌交换，相当于对牌堆做了多轮随机扰动，每位玩家发13张牌，使用读写锁保证并发安全。发牌时每个玩家只能看到自己的手牌，其他玩家手牌显示为隐藏状态
- 码牌：牌的编码系统，每张麻将牌都有唯一的数字编码
  - 万(1-9)、筒(11-19)、条(21-29)、东西南北(31-34)、中(35)

6.**用户B能不能判断用户A胡牌，为什么没有设置机器人打**

- 不能，胡牌判断完全在服务端进行，采用查表法实现，客户端只能看到操作结果，无法获取其他玩家的手牌信息，确保游戏公平性。
- 项目重点在于核心游戏逻辑和分布式架构实现，机器人AI需要额外的策略算法开发。当前支持托管功能，玩家可设置自动操作。
  - 托管：30秒内没有出牌，超时之后自动出牌
  - 出牌决策：
    - 优先执行弃牌操作（选择当前摸到的牌进行弃牌）
    - 如果没有弃牌选项，则执行过牌操作

7.**有没有线下测试过，怎么测试的**

- 有测试机制：
  - 支持手动指定测试牌进行调试
    - 系统维护testCardArray数组存储客户端传来的指定的测试牌，发牌时优先从测试牌数组获取，使用后将对应位置清零
  - 集成监控和日志系统便于问题排查
    - 基于charmbracelet/log库封装，支持多个日志级别：DEBUG、INFO、WARN、ERROR、FATAL
    - 通过配置文件控制日志级别，DEBUG模式显示详细信息
- 单元测试 ：对胡牌算法进行测试验证
  - 创建不同牌型组合的测试用例；验证胡牌判断逻辑的正确性；测试鬼牌替换机制
- 功能测试 ：使用测试牌功能验证游戏逻辑
  - 通过 testCardArray 设置特定牌型；验证麻将胡牌算法和比牌逻辑；测试各种游戏场景的正确性
- 压力测试 ：通过WebSocket连接模拟多用户并发访问
  - 利用连接管理机制；创建大量并发WebSocket连接；测试系统在高并发下的稳定性和性能表现
- 集成测试 ：通过gRPC接口测试各微服务间的协作
  - 使用 `resolver.go` 中的服务发现机制
    - resolver：把客户端传入的目标地址（比如`etcd:///user-service`）解析成一组可用的后端服务器地址，并把这些地址交给 gRPC 的负载均衡器使用
  - 测试etcd注册中心的服务注册与发现
  - 验证负载均衡策略（如round_robin）的有效性

8.**项目有实现****垃圾回收****和****内存****对齐吗**

- 垃圾回收
  - Go语言自带GC
  - 使用对象池模式复用频繁创建的对象（项目中没有）
  - 及时释放不再使用的大对象引用
- 内存对齐
  - Go编译器自动处理内存对齐
  - 结构体字段按照大小排序，减少内存碎片
  - 使用sync.Pool复用临时对象，减少内存分配（项目中没有）

9.**Gin如何鉴权**

- 使用JWT Token进行鉴权
  - 用户注册/登录后生成JWT token -> 客户端在请求头中携带token -> 服务端验证token有效性 -> 验证通过后提取用户ID进行后续操作

10.**JWT****为什么安全，是****对称加密****还是非对称**

- 项目使用对称加密 （HMAC-SHA256）：
  - 结构 ：Header.Payload.Signature三部分
  - 安全性 ：使用密钥签名防止token被篡改，服务端验证确保Token完整性
  - 优势 ：无状态（服务端不需要存储session）、跨服务（多个服务实例共享验证逻辑）、性能好
  - 注意 ：需要保护好Secret密钥，设置合理过期时间

11.**高可用服务怎么构建**

- 多实例部署，单点故障自动切换
- Etcd租约机制，服务异常自动摘除
- 熔断降级 ：超时机制和错误处理
  - 熔断降级是分布式系统的容错机制：当某个服务出现故障或响应时间过长时，熔断器会"断开"对该服务的调用，直接返回预设的降级响应（返回缓存的历史数据或静态配置数据/返回空数据但保持接口可用），避免故障传播导致整个系统崩溃。
- 集成metrics监控服务状态
- MongoDB保证数据持久化，Redis缓存提升数据访问性能并支持持久化防止缓存数据丢失

12.**项目中****负载均衡****的实现**

- 基于etcd服务发现获取服务实例列表，使用gRPC的round_robin策略进行轮询负载均衡，支持服务权重配置，通过 `selectDst` 方法随机选择目标服务器，实现了请求的均匀分发。

13**并发上遇到的数据一致性的问题**

- 游戏状态同步 → 读写锁保护 ：使用sync.RWMutex保护共享的游戏状态数据
- 用户数据更新 → 数据库事务 ：通过MongoDB事务保证数据更新的原子性
- 房间管理 → 消息队列顺序 ：NATS通过主题订阅保证同一房间消息的顺序处理
  - NATS通过单一订阅者模式和消息确认机制保证操作顺序，同一主题的消息按发送顺序依次处理。
- 计数器操作 → 原子操作 ：使用atomic包处理简单的计数器更新

14.**讲一讲****剪枝**

胡牌算法优化方面：

- 在递归生成胡牌组合时，提前判断不可能的分支
- 通过牌数统计避免无效的递归
- 预计算所有可能的胡牌组合，将计算结果存储在内存查找表中，避免每次胡牌判断时的重复递归计算

15.**项目的消息路由**

- 路由格式 ： serverType.handler.method （如 hall.roomHandler.joinRoom ）
  - 服务类型：标识目标服务的类型，用于服务发现和负载均衡
  - 处理器：标识具体的业务处理器类，按功能模块划分
  - 方法：标识具体的业务方法，对应实际的处理函数
- 本地路由 ：首先检查是否为本地connector服务处理
- 远程路由 ：通过 `selectDst` 选择目标服务器，使用NATS发送到对应微服务
- 消息封装 ：将请求封装为 `Msg` 结构，包含源服务器、目标服务器、路由信息等
- 响应处理 ：目标服务处理完成后，通过相同路径返回响应给客户端

16.**消息队列****怎么保证数据不丢**

- 消息持久化，防止服务重启导致消息丢失
- 消息确认机制，确保消息被正确处理
- 重试机制，处理临时性失败
- 死信队列，处理无法正常处理的消息
  - 死信队列是消息队列系统中用于存储无法正常处理消息的特殊队列。当消息出现以下情况时会被发送到死信队列（消息进入 DLQ 后可以：重试、人工处理、记录分析或丢弃）：
    - 处理失败 ：消息处理过程中发生异常且重试次数超限
    - 过期消息 ：消息在队列中停留时间超过TTL（生存时间）
    - 队列满载 ：目标队列已满无法接收新消息
    - 路由失败 ：消息无法找到合适的消费者或路由目标
- NATS：
  - 发布确认机制 ：发送消息返回状态确认发送结果；发送失败时记录错误日志，上层可实现重试机制
  - 订阅确认机制 ：建立可靠订阅，确保消息能够被正确接收，消息接收后立即放入本地通道，避免处理过程中丢失
  - 连接保活机制 ：通过连接状态监控，确保NATS连接的稳定性，连接断开时自动重连，保证消息传输的连续性

17.**Pomelo**

- Pomelo协议是 Pomelo 框架中自定义的一套客户端与服务器通信协议，基于 TCP/WebSocket，结合路由机制 + protobuf/JSON 编解码，为分布式实时游戏和应用提供高效通信能力。
- Pomelo协议在项目中作为WebSocket通信的消息格式标准，定义了完整的消息编解码规范：支持Request、Response、Notify、Push四种消息类型，实现了高效的二进制消息传输，提供了路由压缩、心跳检测等优化机制，确保了客户端与服务器间的可靠实时通信，是整个游戏通信架构的核心协议基础。

18.**protobuf优势**

- 性能优异 ：二进制序列化，比JSON小3-10倍，更紧凑
- 跨语言 ：支持多种编程语言
- 向后兼容 ：字段可选，版本升级友好
- 类型安全 ：强类型定义，编译时类型验证，减少运行时错误
- 代码生成 ：自动生成序列化代码