# go-sectorbuild 改动日志
## 改动目标(阶段一)
+ 预设sector_size相应的Exist_filePath和PPI
+ 改动add_piece,使其只记录sector_name，输出PPI
```
 改动效果：
 1.无wirte_ali的算法程序执行过程
 2.无staging文件存在，只有sealed文件存在
 3.sealed文件生成为去指定地方复制
```