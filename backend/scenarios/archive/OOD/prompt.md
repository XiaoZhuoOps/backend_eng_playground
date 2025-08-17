请根据原文内容 撰写一个英文版本的LLM prompt

# 要求：
0. 你是一个了解LLM Prompt的专家
1. 你对原文内容相关领域有非常深入的研究 取得过对应的博士学位 
2. 首先 深入理解原文的内容 最终形成清晰的Goals
3. 然后 根据你的领域知识 补充必要的SubTopic, Constrains 和 In-Depth Focus
4. 最后 只需要返回结构化的Prompt 控制2k个字符内 放入一个单独的markdown中 不需要具体实现

# 原文:
```
# Context
过期实现

淘汰策略

- FIFO
- LRU LFU

存储数据类型

- 字节
- 对象

底层实现

- ringbuf
- map

Others

- singleflight

# Role
你是一个软件工程领域的专家和计算机科学教授

# Task
参考上述topic，结合与local cache相关的其他topic
撰写一份技术报告，会分为多个step完成
- Step1: 以图形化方式，全面系统地介绍local cache的各个topic，要求简单便于记忆

- Step2: 深入解释每个topic，必须带有图表，文档或论文链接
- Step3: 回答几个关键问题
	- ringbuf底层原理，为什么其GC成本较小
	- GC是如何消耗CPU的，为什么map类型缓存GC成本更高
- Step4: 给出golang代码，以UT的方式评估工业界常用的几种localcache的性能，尤其是GC，序列化等
```