#### 表格

```
CREATE TABLE `metric` (
  `Minute` UInt32,
  `MetricId` FixedString(36),
  `N` Uint32
) ENGINE = MergeTree()
ORDER BY (Minute,MetricId)

```

#### 测试

100万个metric，每个metric每分钟插入10条数据，测试插入与查询性能

1. 插入，共1000万行

   ~~~
   1C/8G docker，插入3列、10000000行，prepare 16秒, commit 2秒
   ~~~

2. 查询，计算每个metric在某一分钟N的sum，min，max，avg

   ~~~sql
   select sum(N),max(N),min(N),avg(N) 
   from metric 
   where Minute=26061293 group by MetricId
   ~~~

   结果

   ~~~
   1000000 rows in set. Elapsed: 1.525 sec. Processed 10.00 million rows, 440.10 MB (6.56 million rows/s., 288.59 MB/s.)
   ~~~

