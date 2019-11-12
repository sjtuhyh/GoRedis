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
redis 采用网络IO多路复用技术来保证在多连接的时候， 系统的高吞吐量。

多路-指的是多个socket连接，复用-指的是复用一个线程。多路复用主要有三种技术：select，poll，epoll。epoll是最新的也是目前最好的多路复用技术。

多路复用是指使用一个线程来检查多个文件描述符（Socket）的就绪状态，比如调用select和poll函数，传入多个文件描述符（FileDescription，简称FD），
如果有一个文件描述符（FileDescription）就绪，则返回，否则阻塞直到超时。得到就绪状态后进行真正的操作可以在同一个线程里执行，也可以启动线程执行（比如使用线程池）。
就是派一个代表，同时监听多个文件描述符是否有数据到来。等着等着，如有有数据，就告诉某某你的数据来啦！赶紧来处理吧。

![image](https://user-images.githubusercontent.com/34932312/68671728-f444d980-058a-11ea-84b2-fd65f809ae0b.png)

### select

### poll

### epoll