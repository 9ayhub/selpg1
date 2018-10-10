## 1.设计说明

### 1.1概述

CLI（Command Line Interface）实用程序是Linux下应用开发的基础。在开发领域，CLI在编程、调试、运维、管理中提供了图形化程序不可替代的灵活性与效率。

用go语言实现了“开发 Linux 命令行实用程序”一文中已经用c语言实现了的selpg程序。

### 1.2代码注释

代码中添加了很多注释，希望能帮助理解。

- **///////selpg_args struct///////开始的部分**

  这部分的变量命名与功能与原代码基本一致。
  
  这部分定义了一个类型，即selpg_args 结构。事实上，这个结构存储了用户从命令行输入的一些信息，比如输入：
  ```html
  $ selpg -s 1 -e 1 test.txt
  ```
  selpg_args里就将start_page和end_page设为1，将in_filename设为test.txt，其他参数均为默认值。
  

- **////////main///////// 开始的部分**

  由于我们使用pflag绑定了sa的各个变量，我们可以省略在main初始化的部分，其他基本一致。
  
  main函数首先声明一个名为sa的selpg_args，接着调用其他两个函数。
  

- **//////func process_args//////// 开始的部分**

  这部分使用Pflag来帮助我们分析参数，但仍然进行了必要的错误检查。
  
  process_args函数主要是分析用户输入的命令，进行错误检查，并将各种信息存储在sa中。
  

- **/////func process_input/////// 开始的部分**

  与selpg.c一样，我们先选择从哪里读取和在哪儿打印，接着按照page_type进行打印。不同的是，当用户指定了输出地点时，我们通过cmd创建子程序“cat”，
帮助我们将输出流的内容打印到指定地点。同时，这部分也是整个代码最难理解的部分。


## 2.使用


命令行格式如下：

```html
selpg -s startPage -e endPage [-l linePerPage | -f ][-d dest] filename
```

其中，-s表示开始打印的页码，-e表示结束打印的页码，这两个必须写上；
而-l表示按固定行数打印文件，-f表示按照换页符来打印，默认按行；-d则是打印的目的地，默认为屏幕。

使用例子：
>
>**1. selpg -s 1 -e 1 input_file**
>
>该命令将把“input_file”的第 1 页写至标准输出（也就是屏幕），因为这里没有重定向或管道。


>**2. other_command | selpg -s10 -e20**
>
>“other_command”的标准输出被 shell／内核重定向至 selpg 的标准输入。将第 10 页到第 20 页写至 selpg 的标准输出（屏幕）。


>**3.selpg -s10 -e20 input_file >output_file 2>error_file**
>
>selpg 将第 10 页到第 20 页写至标准输出，标准输出被重定向至“output_file”；selpg 写至标准错误的所有内容都被重定向至“error_file”。
>当“input_file”很大时可使用这种调用；您不会想坐在那里等着 selpg 完成工作，并且您希望对输出和错误都进行保存。


>**4.selpg -s10 -e20 -l66 input_file**
>
>该命令将页长设置为 66 行，这样 selpg 就可以把输入当作被定界为该长度的页那样处理。第 10 页到第 20 页被写至 selpg 的标准输出（屏幕）。


## 3.测试结果
**1.selpg -s1 -e1 input_file**
**2.selpg -s1 -e1 < input_file**
**3.other_command | selpg -s10 -e20**
**4.selpg -s10 -e20 input_file >output_file**
**5.selpg -s10 -e20 input_file 2>error_file**
**6.selpg -s10 -e20 input_file >output_file 2>error_file**
**7.selpg -s10 -e20 input_file >output_file 2>/dev/null**
**8.selpg -s10 -e20 input_file >/dev/null**
**9.selpg -s10 -e20 input_file | other_command**
**10.selpg -s10 -e20 input_file 2>error_file | other_command**
**11.selpg -s10 -e20 -l66 input_file**
**12.selpg -s10 -e20 -f input_file**
**13.selpg -s10 -e20 -dlp1 input_file**
