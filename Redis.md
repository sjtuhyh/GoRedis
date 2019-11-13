<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [线程IO模型](#%E7%BA%BF%E7%A8%8Bio%E6%A8%A1%E5%9E%8B)
  - [阻塞IO](#%E9%98%BB%E5%A1%9Eio)
  - [非阻塞IO](#%E9%9D%9E%E9%98%BB%E5%A1%9Eio)
  - [IO多路复用](#io%E5%A4%9A%E8%B7%AF%E5%A4%8D%E7%94%A8)
    - [select](#select)
    - [poll](#poll)
    - [epoll](#epoll)
  - [定时任务](#%E5%AE%9A%E6%97%B6%E4%BB%BB%E5%8A%A1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# 线程IO模型
Redis是一个单线程程序！

除了Redis之外，Node.js、Nginx也是单线程，它们都是服务器高性能的典范

Redis单线程为什么还这么快？

1、所有数据都存于内存中，所有的运算都是内存级别的运算  2、redis是单线程的，省去了很多上下文切换线程的时间（上下文切换指的是cpu从一个进程(线程)切换到另一个进程(线程)）
3、redis使用IO多路复用技术来处理并发的连接。非阻塞IO内部实现采用epoll，采用了epoll+自己实现的简单的事件框架。epoll中的读、写、关闭、连接都转化成了事件，然后利用epoll的多路复用特性，绝不在io上浪费一点时间。

## 阻塞IO
当我们调用套接字的读写方法，默认它们是阻塞的，比如read方法要传递进去一个参数n，表示读取这么多字节后再返回，如果没有读够线程就会卡在那里，
直到新的数据到来或者连接关闭了，read 方法才可以返回，线程才能继续处理。而 write 方法一般来说不会阻塞，除非内核为套接字分配的写缓冲区已经满了，write方法就会阻塞，直到缓存区中有空闲空间挪出来了。
![image](https://user-images.githubusercontent.com/34932312/68669740-39b2d800-0586-11ea-9be8-7e024123a7a1.png)

## 非阻塞IO
非阻塞IO很简单，通过fcntl（POSIX）或ioctl（Unix）设为非阻塞模式，这时，当你调用read时，如果有数据收到，就返回数据，如果没有数据收到，就立刻返回一个错误，
如EWOULDBLOCK。这样是不会阻塞线程了，但是你还是要不断的轮询来读取或写入。相当于你去查看有没有数据，告诉你没有，过一会再来吧！应用过一会再来问，有没有数据？没有数据，会有一个返回。
采用死循环方式轮询每一个流，如果有IO事件就处理，这样可以使得一个线程可以处理多个流，但是效率不高，容易导致CPU空转，应用必须得过一会来一下，问问内核有木有数据啊。这和现实很像啊！好多情况都得去某些地方问问好了没有？木有，明天再过来。明天，好了木有？木有，后天再过来。。。。。忙碌的应用。。。。
![image](https://user-images.githubusercontent.com/34932312/68670051-ea20dc00-0586-11ea-93ab-90b39ecec8b6.png)

## IO多路复用
redis 采用网络IO多路复用技术来保证在多连接的时候，系统的高吞吐量。

多路-指的是多个socket连接，复用-指的是复用一个线程。多路复用主要有三种技术：select，poll，epoll。epoll是最新的也是目前最好的多路复用技术。

多路复用是指使用一个线程来检查多个文件描述符（Socket）的就绪状态，比如调用select和poll函数，传入多个文件描述符（FileDescription，简称FD），
如果有一个文件描述符（FileDescription）就绪，则返回，否则阻塞直到超时。得到就绪状态后进行真正的操作可以在同一个线程里执行，也可以启动线程执行（比如使用线程池）。
就是派一个代表，同时监听多个文件描述符是否有数据到来。等着等着，如有有数据，就告诉某某你的数据来啦！赶紧来处理吧。

目前支持I/O多路复用的系统调用有 select，pselect，poll，epoll，I/O多路复用就是通过一种机制，一个进程可以监视多个描述符，一旦某个描述符就绪（一般是读就绪或者写就绪），能够通知程序进行相应的读写操作。但select，pselect，poll，epoll本质上都是同步I/O，因为他们都需要在读写事件就绪后自己负责进行读写，也就是说这个读写过程是阻塞的，而异步I/O则无需自己负责进行读写，异步I/O的实现会负责把数据从内核拷贝到用户空间。

![image](https://user-images.githubusercontent.com/34932312/68671728-f444d980-058a-11ea-84b2-fd65f809ae0b.png)

### select
基本原理：select函数监视的文件描述符分3类，分别是writefds、readfds、和exceptfds。
调用后select函数会阻塞，直到有描述符就绪（有数据可读、可写、或者有except），或者超时（timeout指定等待时间，如果立即返回设为null即可），
函数返回。当select函数返回后，可以通过遍历fdset，来找到就绪的描述符。
     
基本流程：
![image](https://user-images.githubusercontent.com/34932312/68765062-e197d680-0656-11ea-968e-b62b5aca086d.png)
select目前几乎在所有的平台上支持，其良好跨平台支持也是它的一个优点。select的一个缺点在于单个进程能够监视的文件描述符的数量存在最大限制，
在Linux上一般为1024，可以通过修改宏定义甚至重新编译内核的方式提升这一限制，但是这样也会造成效率的降低。

select本质上是通过设置或者检查存放fd标志位的数据结构来进行下一步处理。这样所带来的缺点是：

1、select最大的缺陷就是单个进程所打开的FD是有一定限制的，它由FD_SETSIZE设置，默认值是1024。

2、对socket进行扫描时是线性扫描，即采用轮询的方法，效率较低。当套接字比较多的时候，每次select()都要通过遍历FD_SETSIZE个Socket来完成调度，不管哪个Socket是活跃的，都遍历一遍。这会浪费很多CPU时间。如果能给套接字注册某个回调函数，当他们活跃时，自动完成相关操作，那就避免了轮询，这正是epoll与kqueue做的。
                                  
3、需要维护一个用来存放大量fd的数据结构，每次调用select，都需要把fd集合从用户态拷贝到内核态，这个开销在fd很多时会很大，同时每次调用select都需要在内核遍历传递进来的所有fd，这个开销在fd很多时也很大

伪代码：
输入读写描述符列表 read_fds & write_fds，输出与之对应的可读可写事件。同时还提供了一个 timeout 参数，如果没有任何事件到来，那么就最多等待 timeout 时间，线程处于阻塞状态。
```
read_events, write_events = select(read_fds, write_fds, timeout)
for event in read_events:
    handle_read(event.fd)
for event in write_events:
    handle_write(event.fd)
handle_others() # 处理其它事情，如定时任务等
```
### poll

poll本质上和select没有区别，它将用户传入的数组拷贝到内核空间，然后查询每个fd对应的设备状态，但是它没有最大连接数的限制，原因是它是基于链表来存储的.

### epoll
https://www.jianshu.com/p/dfd940e7fca2

## 定时任务
Redis 的定时任务会记录在一个称为最小堆的数据结构中。这个堆中，最快要执行的任务排在堆的最上方。在每个循环周期，Redis 都会将最小堆里面已经到点的任务立即进行处理。处理完毕后，将最快要执行的任务还需要的时间记录下来，这个时间就是 select 系统调 用的 timeout 参数。因为 Redis 知道未来 timeout 时间内，没有其它定时任务需要处理，所以 可以安心睡眠 timeout 的时间。
