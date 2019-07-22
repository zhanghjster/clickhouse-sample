#### 简介

梳理clickhouse基本资料，体验写入、查询性能

#### 资料

##### 介绍

clickhouse是用于联机分析(OLAP)的列式数据库管理系统。

* 行式数据库，同一行的数据在物理上存储在一起，要读取某列数据需要读取整行

  mysql、postgre 属于行式存储

* 列式数据库，同一列的数据在物理上存在一起，不同列数据分开存储

不同存储方式适用于不同场景，这些场景包括查询种类、频率、数据量、延迟，写入与查询关系、事务、完整性等。没有一种系统适用于明显不同的场景。

Clickhouse适用典型的OLAP场景

* 写多读少，以大批量(>1000rows)写入，不修改数据

* 宽表，有大量列，列数据比较小(数字或短字符)
* 查询读取大量行，少量列
* 少量查询，每台机器每秒几百甚至更少
* 查询允许有延迟(>50ms)，结果明显小于数据源(聚合操作)
* 不要求事务、一致性要求低

原因

* 分析类查询每次只需要读取少量列，每列分开存储则每次只读出需要的数。比如100列中读出5列，则可减少20倍IO

* 数据被批量读取的方式使压缩更容易，较少了IO也更容易使用系统缓存

   <img src="https://clickhouse.yandex/docs/zh/images/row_oriented.gif" width="400">行式

  <img src="https://clickhouse.yandex/docs/zh/images/column_oriented.gif" width="400"> 列式 



##### 特性

1. 真正的列式存储，列中除了数据本身外不存储其他数据，数据更紧凑更有利于读写新能(缩减IO、高效压缩)。相比于Hbase、Cassandra等列式存储每秒几十万的吞吐能力，Clickhouse可以达到每秒几亿行的吞吐能力
2. 数据压缩
3. 数据磁盘存储，被设计应用在传统磁盘上，更低的存储成本
4. 多核并行处理，大型查询使用并行化进行
5. 多服务器分布式处理，数据保存在不同的shard上，每个shard由一组容错的replca组成，查询并行的在不同shard上进行
6. 支持SQL对数据库进行管理以及对数据的增删改查
7. 向量引擎，索引，近似计算
8. 支持复制，异步多主复制
9. 不支持事务、缺少对高频低延迟修改的支持、不适合检索单行的点查询

##### 数据类型

* UInt8|16|32|64，Int8|16|32|64|，Float32|64，Decimal，Bool，String，FixedString
* Date，两字节存储1970-01-01到当前时间的天数，写入查询格式是'yyyy-mm-dd'，没有存储时区
* DateTime，四字节存储Unix时间戳，精确度到秒，使用客户端或服务器的系统时区
* Enum，Array，NestedTuple
* AggregateFunction和其他不常用类型

##### 引擎类型

* MergeTree

  clickhouse最强大的引擎，适用于分批大量写入场景

  * 数据按照主键排序，用主键实现快速检索
  * 使用分区，检索条件里包含分区字段的情况下会加速查询。类似分表
  * 支持副本
  * 支持采样

  ~~~sql
  CREATE TABLE [IF NOT EXISTS] [db.]table_name [ON CLUSTER cluster]
  (
      name1 [type1] [DEFAULT|MATERIALIZED|ALIAS expr1],
      name2 [type2] [DEFAULT|MATERIALIZED|ALIAS expr2],
      ...
      INDEX index_name1 expr1 TYPE type1(...) GRANULARITY value1,
      INDEX index_name2 expr2 TYPE type2(...) GRANULARITY value2
  ) ENGINE = MergeTree()
  [PARTITION BY expr]
  [ORDER BY expr]
  [PRIMARY KEY expr]
  [SAMPLE BY expr]
  [SETTINGS name=value, ...]
  ~~~

  * partition by，分区字段，partition by (clumn1, column2), 分区键要谨慎选择，不宜分的过于精细，比如超过1000个分区
  * order by，排序字段
  * primary key，主键，不设置时候为和order by字段相同
  * sample by，取样键

* 其他

#### 测试

环境 2C/4G Docker

##### 表格

~~~~sql
CREATE TABLE `metric` (
  `Minute` UInt32,
  `MetricId` FixedString(36),
  `N` Uint32
) ENGINE = MergeTree()
ORDER BY (Minute,MetricId)
~~~~

##### 写入

10万个metric，每个metric每分钟10条数据

~~~
10000000行，prepare 16秒, commit 2秒
~~~

##### 查询 

计算每个metric在某一分钟N的sum，min，max，avg

~~~sql
select sum(N),max(N),min(N),avg(N) 
from metric 
where Minute=26061293 group by MetricId
~~~

结果

~~~
1000000 rows in set. Elapsed: 1.525 sec. Processed 10.00 million rows, 440.10 MB (6.56 million rows/s., 288.59 MB/s.)
~~~

##### 对比

ES 10000个Metric，10万条数据做聚合计算